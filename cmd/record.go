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
 * @file record.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package cmd

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"valet/log/logfacade"
	"valet/valet"
)

const defaultSweep = 5 * time.Minute
const minSweep = 30 * time.Second

var recordCmd = &cobra.Command{
	Use: "record",
	Short: "Record checksums under a root directory",
	Long: `
valet record will monitor a directory hierarchy and locate data files within it
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
	Run: runRecordCmd,
}

func init() {
	recordCmd.Flags().StringVarP(&allCliFlags.rootDir, "root",
		"r", "", "root directory to search")
	err := recordCmd.MarkFlagRequired("root")
	if err != nil {
		logfacade.GetLogger().Error().
			Err(err).Msg("failed to mark --root required")
		os.Exit(1)
	}

	recordCmd.Flags().DurationVarP(&allCliFlags.sweepInterval, "interval",
		"i", defaultSweep, "directory sweep interval, minimum 30s")

	valetCmd.AddCommand(recordCmd)
}

func runRecordCmd(cmd *cobra.Command, args []string) {
	log := setupLogger(allCliFlags)
	root := allCliFlags.rootDir
	interval := allCliFlags.sweepInterval
	pred := valet.RequiresChecksum

	cancelCtx, _ := context.WithCancel(context.Background())

	wpaths, werrs := valet.WatchFiles(cancelCtx, root, pred)
	fpaths, ferrs := valet.FindFilesInterval(cancelCtx, root, pred, interval)
	mpaths := mergeFileChannels(wpaths, fpaths)
	errs := mergeErrorChannels(werrs, ferrs)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := valet.ProcessFiles(mpaths, valet.RecordChecksum, allCliFlags.maxProc)
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

func mergeFileChannels(x <-chan valet.FilePath, y <-chan valet.FilePath) chan valet.FilePath {
	merged := make(chan valet.FilePath)

	log := logfacade.GetLogger()

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
