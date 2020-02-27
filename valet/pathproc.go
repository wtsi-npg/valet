/*
 * Copyright (C) 2019, 2020. Genome Research Ltd. All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * @file pathproc.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"context"
	"sync"
	"time"

	logs "github.com/kjsanger/logshim"
	"github.com/pkg/errors"
)

type token struct{}
type semaphore chan token

type ProcessParams struct {
	Root          string        // The local root directory to work on.
	MatchFunc     FilePredicate // The file selecting predicate.
	PruneFunc     FilePredicate // The local directory tree pruning predicate.
	Plan          WorkPlan      // The plan for selected files.
	SweepInterval time.Duration // The interval between sweeps of the local directory tree.
	MaxProc       int           // The maximum number of threads to run.
}

// ProcessFiles detects files to work on, dispatches any files found to
// suitable work functions and monitors any errors that occur during the
// detection and processing steps. The function will continue to run until
// cancelled.
//
// Errors that occur in detection are logged as warnings, but do not cause this
// function to return an error itself. Error that occur during processing are
// counted. If when cancelled, this function has counted any processing errors,
// it will return an error itself.
func ProcessFiles(cancelCtx context.Context, params ProcessParams) error {

	wpaths, werrs := WatchFiles(cancelCtx, params.Root, params.MatchFunc,
		params.PruneFunc)
	fpaths, ferrs := FindFilesInterval(cancelCtx, params.Root, params.MatchFunc,
		params.PruneFunc, params.SweepInterval)

	paths := MergeFileChannels(wpaths, fpaths)
	errs := MergeErrorChannels(werrs, ferrs)
	log := logs.GetLogger()

	// Inform the user that cancellation has started because it can take a
	// while for jobs to complete. This blocks until then, or until the data-
	// and error-processing goroutines return, when a send on the noCancelMsg
	// channel allows this goroutine to return.
	noCancelMsg := make(chan token, 1)
	go func() {
		select {
		case <-noCancelMsg:
			return
		case <-cancelCtx.Done():
			log.Info().Msg("processing cancelled, waiting for jobs")
			return
		}
	}()

	wg := sync.WaitGroup{}
	var perr error

	wg.Add(1)
	go func() {
		defer wg.Done()

		perr = DoProcessFiles(paths, params.Plan, params.MaxProc)
	}()

	// Log as warnings any errors encountered
	wg.Add(1)
	go func() {
		defer wg.Done()

		for err := range errs {
			log.Warn().Err(err).Msg("while detecting files")
		}
	}()

	wg.Wait()

	noCancelMsg <- token{}
	log.Info().Msg("processing done")

	return perr
}

// DoProcessFiles operates by applying workPlan to each FilePath in the paths
// channel. Each WorkPlan is executed in its own goroutine, with no more than
// maxThreads goroutines running in parallel.
//
// This function keeps track of the FilePaths being worked on. If a FilePath is
// passed in subsequently, but before existing work has finished, it is skipped.
//
// If any WorkPlan encounters an error, the error is logged and counted. When
// DoProcessFiles exits, it will return an error if the error count across all
// the WorkPlans was greater than 0.
func DoProcessFiles(paths <-chan FilePath, workPlan WorkPlan, maxThreads int) error {
	var wg sync.WaitGroup // The group of all work goroutines

	var mu = sync.Mutex{} // Protects running, jobCount, errCount
	var running = make(map[string]token)
	var jobCount uint64
	var errCount uint64

	sem := make(semaphore, maxThreads) // Ensure upper limit on thread count

	log := logs.GetLogger()

	for path := range paths {
		mu.Lock()
		if _, ok := running[path.Location]; ok {
			mu.Unlock()
			log.Info().Str("path", path.Location).
				Msg("skipping (already working on)")
			continue
		}
		mu.Unlock()

		sem <- token{}
		wg.Add(1)

		go func(p FilePath) {
			defer func() {
				<-sem
				wg.Done()
			}()

			mu.Lock()
			running[p.Location] = token{}
			jobCount++

			work, derr := makeWork(p, workPlan)
			if derr != nil {
				mu.Unlock()
				log.Error().Err(derr).
					Str("path", p.Location).
					Msg("work dispatch failed")
				errCount++
				return
			}
			mu.Unlock()

			log.Debug().Str("path", p.Location).
				Str("plan", workPlan.String()).Msg("starting work")
			werr := work.WorkFunc(p)
			log.Debug().Str("path", p.Location).
				Str("plan", workPlan.String()).Msg("finished work")

			mu.Lock()
			delete(running, p.Location)
			if werr != nil {
				errCount++
				mu.Unlock()
				log.Error().Err(werr).
					Str("path", p.Location).
					Msg("worker function failed")
				return
			}
			mu.Unlock()
		}(path)
	}

	wg.Wait()

	if errCount > 0 {
		return errors.Errorf("encountered %d errors processing %d files",
			errCount, jobCount)
	}

	log.Info().Uint64("num_files", jobCount).Msg("finished processing")

	return nil
}
