/*
 * Copyright (C) 2019. Genome Research Ltd. All rights reserved.
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
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	logf "valet/log/logfacade"
)

type token struct{}
type semaphore chan token

// ProcessFiles operates by applying workFunc to each FilePath in the paths
// channel. An error is raised if any workFunc itself encounters an error.
func ProcessFiles(paths <-chan FilePath, workPlan WorkPlan, maxThreads int) error {
	var wg sync.WaitGroup
	var jobCounter uint64
	var errCounter uint64
	sem := make(semaphore, maxThreads)

	log := logf.GetLogger()

	for path := range paths {
		wg.Add(1)
		sem <- token{}

		go func(p FilePath) {
			defer func() {
				<-sem
				wg.Done()
			}()

			atomic.AddUint64(&jobCounter, 1)
			log.Info().Str("path", p.Location).Msg("working on")

			work, derr := DispatchWork(p, workPlan)
			if derr != nil {
				log.Error().Err(derr).
					Str("path", p.Location).
					Msg("work dispatch failed")
				atomic.AddUint64(&errCounter, 1)
				return
			}

			werr := work.WorkFunc(p)
			if werr != nil {
				log.Error().Err(werr).
					Str("path", p.Location).
					Msg("worker function failed")
				atomic.AddUint64(&errCounter, 1)
			}
		}(path)
	}

	wg.Wait()

	jobCount := atomic.LoadUint64(&jobCounter)
	errCount := atomic.LoadUint64(&errCounter)
	if errCount > 0 {
		return errors.Errorf("encountered %d errors processing %d files",
			errCount, jobCount)
	}

	log.Info().Uint64("num_files", jobCount).Msg("finished processing")

	return nil
}
