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
 * @file checksum_create.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package cmd

import (
	"context"
	"os"
	"time"

	logs "github.com/kjsanger/logshim"
	"github.com/spf13/cobra"

	"github.com/kjsanger/valet/valet"
)

var checksumCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create checksum files under a root directory",
	Long: `
valet checksum create will monitor a directory hierarchy and locate data files
within it that have no accompanying checksum file, or have a checksum file that
is stale. valet will then calculate the checksum and create or update the
checksum file.

- Creating up-to-date checksum files
  
  - Directory hierarchy styles supported

    - Any
  
  - File patterns supported

    - All supported for archiving

  - Checksum file patterns supported

    - (data file name).md5
`,
	Example: `
valet checksum create --root /data --exclude /data/intermediate \
    --exclude /data/queued_reads --exclude /data/reports \
    --interval 20m --verbose`,
	Run: runChecksumCreateCmd,
}

func init() {
	checksumCreateCmd.Flags().StringVarP(&allCliFlags.localRoot,
		"root", "r", "",
		"the root directory of the monitor")

	err := checksumCreateCmd.MarkFlagRequired("root")
	if err != nil {
		logs.GetLogger().Error().
			Err(err).Msg("failed to mark --root required")
		os.Exit(1)
	}

	checksumCreateCmd.Flags().DurationVarP(&allCliFlags.sweepInterval,
		"interval", "i", valet.DefaultSweep,
		"directory sweep interval, minimum 30s")

	checksumCreateCmd.Flags().BoolVar(&allCliFlags.dryRun,
		"dry-run", false,
		"dry-run (make no changes)")

	checksumCreateCmd.Flags().StringArrayVar(&allCliFlags.excludeDirs,
		"exclude", []string{},
		"patterns matching directories to prune "+
			"from both monitoring and interval sweeps")

	checksumCmd.AddCommand(checksumCreateCmd)
}

func runChecksumCreateCmd(cmd *cobra.Command, args []string) {
	log := setupLogger(allCliFlags)
	root := allCliFlags.localRoot
	exclude := allCliFlags.excludeDirs
	interval := allCliFlags.sweepInterval
	maxProc := allCliFlags.maxProc
	dryRun := allCliFlags.dryRun

	if interval < valet.MinSweep {
		log.Error().Msgf("Invalid interval %s (must be > %s)",
			interval, valet.MinSweep)
		os.Exit(1)
	}

	err := CreateChecksumFiles(root, exclude, interval, maxProc, dryRun)

	if err != nil {
		log.Error().Err(err).Msg("checksum creation failed")
		os.Exit(1)
	}
}

func CreateChecksumFiles(root string, exclude []string, interval time.Duration,
	maxProc int, dryRun bool) error {
	log := logs.GetLogger()

	cancelCtx, cancel := context.WithCancel(context.Background())
	setupSignalHandler(cancel)

	// pruneFn, err := valet.MakeRegexPruneFn(exclude)
	pruneFn, err := valet.MakeGlobPruneFunc(exclude)
	if err != nil {
		log.Error().Err(err).Msg("error in exclusion patterns")
		os.Exit(1)
	}

	var workPlan valet.WorkPlan
	if dryRun {
		workPlan = valet.DryRunWorkPlan()
	} else {
		workPlan = valet.CreateChecksumWorkPlan()
	}

	return valet.ProcessFiles(cancelCtx, valet.ProcessParams{
		Root:          root,
		MatchFunc:     valet.RequiresChecksum,
		PruneFunc:     pruneFn,
		Plan:          workPlan,
		SweepInterval: interval,
		MaxProc:       maxProc,
	})
}
