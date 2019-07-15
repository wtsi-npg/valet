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
 * @file checksum.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	logf "valet/log/logfacade"
	"valet/utilities"
	"valet/valet"
)

const defaultSweep = 5 * time.Minute
const minSweep = 30 * time.Second

var checksumCmd = &cobra.Command{
	Use:   "checksum",
	Short: "Record checksums under a root directory",
	Long: `
valet checksum will monitor a directory hierarchy and locate data files within it
that have no accompanying checksum file, or have a checksum file that is stale.
valet will then calculate the checksum and create or update the checksum file.

- Creating up-to-date checksum files
  
  - Directory hierarchy styles supported

    - Any
  
  - File patterns supported

    - *.fast5$
    - *.fastq$

  - Checksum file patterns supported

    - (data file name).md5
`,
	Example: `
valet checksum --root /data --exclude /data/intermediate \
    --exclude /data/queued_reads --exclude /data/reports \
    --interval 20m --verbose`,
	Run: runChecksumCmd,
}

func init() {
	checksumCmd.Flags().StringVarP(&allCliFlags.rootDir, "root",
		"r", "", "the root directory of the monitor")
	err := checksumCmd.MarkFlagRequired("root")
	if err != nil {
		logf.GetLogger().Error().
			Err(err).Msg("failed to mark --root required")
		os.Exit(1)
	}

	checksumCmd.Flags().DurationVarP(&allCliFlags.sweepInterval, "interval",
		"i", defaultSweep, "directory sweep interval, minimum 30s")

	checksumCmd.Flags().BoolVar(&allCliFlags.dryRun, "dry-run", false,
		"dry-run (make no changes)")

	checksumCmd.Flags().StringArrayVar(&allCliFlags.excludeDirs, "exclude",
		[]string{}, "patterns matching directories to prune " +
		"from both monitoring and interval sweeps")

	valetCmd.AddCommand(checksumCmd)
}

func runChecksumCmd(cmd *cobra.Command, args []string) {
	log := setupLogger(allCliFlags)
	root := allCliFlags.rootDir
	interval := allCliFlags.sweepInterval
	pred := valet.RequiresChecksum
	maxProc := allCliFlags.maxProc
	dryRun := allCliFlags.dryRun

	cancelCtx, cancel := context.WithCancel(context.Background())
	setupSignalHandler(cancel)

	// pruneFn, err := makeRegexPruneFn(allCliFlags.excludeDirs)
	pruneFn, err := makeGlobPruneFn(allCliFlags.excludeDirs)
	if err != nil {
		log.Error().Err(err).Msg("error in exclusion patterns")
		os.Exit(1)
	}

	wpaths, werrs := valet.WatchFiles(cancelCtx, root, pred, pruneFn)
	fpaths, ferrs := valet.FindFilesInterval(cancelCtx, root, pred, pruneFn, interval)
	mpaths := mergeFileChannels(wpaths, fpaths)
	errs := mergeErrorChannels(werrs, ferrs)

	var workFn valet.WorkFunc
	if dryRun {
		workFn = valet.DoNothing
	} else {
		workFn = valet.CreateOrUpdateMD5ChecksumFile
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := valet.ProcessFiles(mpaths, workFn, maxProc)
		if err != nil {
			log.Error().Err(err).Msg("failed processing")
			os.Exit(1)
		}
	}()

	if err := <-errs; err != nil {
		log.Error().Err(err).Msg("failed to complete processing")
		os.Exit(1)
	}

	wg.Wait()
}

func setupSignalHandler(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-signals
		log := logf.GetLogger()

		switch s {
		case syscall.SIGINT:
			log.Info().Msg("got SIGINT, shutting down")
			cancel()
		case syscall.SIGTERM:
			log.Info().Msg("got SIGTERM, shutting down")
			cancel()
		default:
			log.Error().Str("signal", s.String()).
				Msg("got unexpected signal, exiting")
			os.Exit(1)
		}
	}()
}

func makeRegexPruneFn(patterns []string) (valet.FilePredicate, error) {
	log := logf.GetLogger()

	var regexes []*regexp.Regexp
	var errors []error
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			errors = append(errors, err)
		} else {
			regexes = append(regexes, re)
		}
	}

	if len(errors) > 0 {
		return nil, utilities.CombineErrors(errors...)
	}

	return func(fp valet.FilePath) (bool, error) {
		for _, regex := range regexes {
			if regex.MatchString(fp.Location) {
				log.Debug().
					Str("path", fp.Location).
					Msg("match path for pruning")
				return true, filepath.SkipDir // return SkipDir to cause walk to skip
			}
		}
		return false, nil
	}, nil
}

func makeGlobPruneFn(patterns []string) (valet.FilePredicate, error) {
	log := logf.GetLogger()

	for _, pattern := range patterns {
		if _, err := filepath.Match(pattern, "."); err != nil {
			return nil, err
		}
	}

	return func(fp valet.FilePath) (bool, error) {
		for _, pattern := range patterns {
			match, err := filepath.Match(pattern, fp.Location)
			if err != nil {
				log.Error().Err(err).Msg("invalid match pattern")
				continue
			}

			if match {
				log.Debug().
					Str("path", fp.Location).
					Msg("matched path for pruning")
				return true, filepath.SkipDir // return SkipDir to cause walk to skip
			}
		}
		return false, nil
	}, nil
}

func mergeFileChannels(x <-chan valet.FilePath, y <-chan valet.FilePath) chan valet.FilePath {
	merged := make(chan valet.FilePath)

	log := logf.GetLogger()

	go func() {
		defer close(merged)

		xOpen, yOpen := true, true
		for xOpen || yOpen {
			select {
			case p, ok := <-x:
				if ok {
					log.Debug().Msg("merging an x path")
					merged <- p
				} else {
					log.Debug().Msg("x was closed")
					xOpen = false
					x = nil
				}

			case p, ok := <-y:
				if ok {
					log.Debug().Msg("merging a y path")
					merged <- p
				} else {
					log.Debug().Msg("y was closed")
					yOpen = false
					y = nil
				}
			}
		}
	}()

	return merged
}

func mergeErrorChannels(x <-chan error, y <-chan error) chan error {
	merged := make(chan error)

	go func() {
		defer close(merged)

		xOpen, yOpen := true, true
		for xOpen || yOpen {
			select {
			case p, ok := <-x:
				if ok {
					merged <- p
				} else {
					xOpen = false
				}

			case p, ok := <-y:
				if ok {
					merged <- p
				} else {
					yOpen = false
				}
			}
		}
	}()

	return merged
}
