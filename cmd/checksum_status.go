/*
 * Copyright (C) 2019, 2020, 2021. Genome Research Ltd. All rights reserved.
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
 * @file checksum_status.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package cmd

import (
	"context"
	"os"
	"sync"

	"github.com/spf13/cobra"
	logs "github.com/wtsi-npg/logshim"

	"github.com/wtsi-npg/valet/valet"
)

var checksumStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check for complete checksum data under a root directory",
	Long: `
valet checksum complete will check for up-to-date checksum files for data files
under a root directory.

`,
	Example: `
valet checksum status --root /data --exclude /data/intermediate \
    --exclude /data/queued_reads --exclude /data/reports \
    --verbose`,
	Run: runChecksumStatusCmd,
}

func init() {
	checksumStatusCmd.Flags().StringVarP(&checksumFlags.localRoot,
		"root", "r", "",
		"the root directory of the monitor")

	err := checksumStatusCmd.MarkFlagRequired("root")
	if err != nil {
		logs.GetLogger().Error().
			Err(err).Msg("failed to mark --root required")
		os.Exit(1)
	}

	checksumStatusCmd.Flags().StringArrayVar(&checksumFlags.excludeDirs,
		"exclude", []string{},
		"patterns matching directories to prune "+
			"from the completeness check")

	checksumCmd.AddCommand(checksumStatusCmd)
}

func runChecksumStatusCmd(cmd *cobra.Command, args []string) {
	log := setupLogger(baseFlags)

	numWithoutChecksum, err :=
	 	CountFilesWithoutChecksum(checksumFlags.localRoot,
			checksumFlags.excludeDirs)
	if err != nil {
		os.Exit(1)
	}

	if numWithoutChecksum == 0 {
		log.Info().Str("root", checksumFlags.localRoot).
			Msg("all checksum files present")
	} else {
		log.Error().Str("root", checksumFlags.localRoot).
			Uint64("count", numWithoutChecksum).
			Msg("checksum files missing")
		os.Exit(1)
	}
}

// CountFilesWithoutChecksum searches for files recursively under root (subject
// to any exclusions patterns in exclude) and counts those that do not have an
// up-to-date checksum file.
func CountFilesWithoutChecksum(root string, exclude []string) (uint64, error) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	setupSignalHandler(cancel)
	log := logs.GetLogger()

	var mu sync.Mutex
	var numWithoutChecksum uint64
	var err error

	pred := valet.RequiresChecksum
	pruneFn, perr := valet.MakeGlobPruneFunc(exclude)
	if perr != nil {
		log.Error().Err(perr).Msg("error in exclusion patterns")
		return numWithoutChecksum, err
	}

	paths, errs := valet.FindFiles(cancelCtx, root, pred, pruneFn)

	countFunc := func(path valet.FilePath) error {
		log.Warn().Str("path", path.Location).Msg("missing checksum")

		mu.Lock()
		numWithoutChecksum++
		mu.Unlock()
		return nil
	}

	maxProcs := 1
	done := make(chan bool)

	go func() {
		defer func() { done <- true }()

		err := valet.DoProcessFiles(paths,
			valet.ChecksumStateWorkPlan(countFunc), maxProcs)
		if err != nil {
			log.Error().Err(err).Msg("failed processing")
			os.Exit(1)
		}
	}()

	<-done

	if err := <-errs; err != nil {
		log.Error().Err(err).Msg("failed to complete processing")
		os.Exit(1)
	}

	return numWithoutChecksum, err
}
