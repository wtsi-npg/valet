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
 * @file checksum_status.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	logf "valet/log/logfacade"
	"valet/valet"
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
	checksumStatusCmd.Flags().StringVarP(&allCliFlags.rootDir,
		"root", "r", "",
		"the root directory of the monitor")

	err := checksumStatusCmd.MarkFlagRequired("root")
	if err != nil {
		logf.GetLogger().Error().
			Err(err).Msg("failed to mark --root required")
		os.Exit(1)
	}

	checksumStatusCmd.Flags().StringArrayVar(&allCliFlags.excludeDirs,
		"exclude", []string{},
		"patterns matching directories to prune "+
			"from the completeness check")

	checksumCmd.AddCommand(checksumStatusCmd)
}

func runChecksumStatusCmd(cmd *cobra.Command, args []string) {
	log := setupLogger(allCliFlags)
	root := allCliFlags.rootDir
	exclude := allCliFlags.excludeDirs

	numWithoutChecksum, err := CountFilesWithoutChecksum(root, exclude)
	if err != nil {
		os.Exit(1)
	}

	if numWithoutChecksum == 0 {
		log.Info().Str("root", root).Msg("all checksum files present")
	} else {
		log.Error().Str("root", root).
			Uint64("count", numWithoutChecksum).
			Msg("checksum files missing")
		os.Exit(1)
	}
}

func CountFilesWithoutChecksum(root string, exclude []string) (uint64, error) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	setupSignalHandler(cancel)
	log := logf.GetLogger()

	var numWithoutChecksum uint64 = 0
	var err error

	pred := valet.RequiresChecksum
	pruneFn, perr := makeGlobPruneFn(exclude)
	if perr != nil {
		log.Error().Err(perr).Msg("error in exclusion patterns")
		return numWithoutChecksum, err
	}

	paths, errs := valet.FindFiles(cancelCtx, root, pred, pruneFn)
	for path := range paths {
		log.Warn().Str("path", path.Location).Msg("without checksum file")
		numWithoutChecksum++
	}

	if err = <-errs; err != nil {
		log.Error().Err(err).Msg("failed to complete processing")
	}

	return numWithoutChecksum, err
}
