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
 * @file archive_create.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package cmd

import (
	"context"
	"os"
	"time"

	ex "github.com/kjsanger/extendo"
	logs "github.com/kjsanger/logshim"
	"github.com/spf13/cobra"

	"github.com/kjsanger/valet/valet"
)

type archiveParams struct {
	exclude     []string
	interval    time.Duration
	maxProc     int
	dryRun      bool
	deleteLocal bool
}

var archiveCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an archive of files under a root directory",
	Long: `
valet archive create will monitor a directory hierarchy and locate data files
within it that are not currently within a remote data store. valet will then
archive them and if successful, delete the archived file from the local disk.

- Archiving files
  
  - Directory hierarchy styles supported

    - Any
  
  - File patterns supported

    - *.fast5$
    - *.fastq$

  - Checksum file patterns supported

    - (data file name).md5
`,
	Example: `
valet archive create --root /data --exclude /data/custom \
    --archive-root /seq/ont/gridion/gxb02004
    --interval 20m --verbose`,
	Run: runArchiveCreateCmd,
}

var maxClients uint8 = 12

func init() {
	archiveCreateCmd.Flags().StringVarP(&allCliFlags.localRoot,
		"root", "r", "",
		"the root directory of the monitor")

	err := archiveCreateCmd.MarkFlagRequired("root")
	if err != nil {
		logs.GetLogger().Error().
			Err(err).Msg("failed to mark --root required")
		os.Exit(1)
	}

	archiveCreateCmd.Flags().StringVarP(&allCliFlags.archiveRoot,
		"archive-root", "a", "",
		"the archive root collection")

	err = archiveCreateCmd.MarkFlagRequired("archive-root")
	if err != nil {
		logs.GetLogger().Error().
			Err(err).Msg("failed to mark --archive-root required")
		os.Exit(1)
	}

	archiveCreateCmd.Flags().DurationVarP(&allCliFlags.sweepInterval,
		"interval", "i", valet.DefaultSweep,
		"directory sweep interval, minimum 30s")

	archiveCreateCmd.Flags().BoolVar(&allCliFlags.dryRun,
		"dry-run", false,
		"dry-run (make no changes)")

	archiveCreateCmd.Flags().StringArrayVar(&allCliFlags.excludeDirs,
		"exclude", []string{},
		"patterns matching directories to prune "+
			"from both monitoring and interval sweeps")

	archiveCreateCmd.Flags().BoolVar(&allCliFlags.deleteLocal,
		"delete-on-archive", false,
		"delete local files on successful archiving")

	archiveCmd.AddCommand(archiveCreateCmd)
}

func runArchiveCreateCmd(cmd *cobra.Command, args []string) {
	log := setupLogger(allCliFlags)
	root := allCliFlags.localRoot
	collection := allCliFlags.archiveRoot

	if allCliFlags.sweepInterval < valet.MinSweep {
		log.Error().Msgf("Invalid interval %s (must be > %s)",
			allCliFlags.sweepInterval, valet.MinSweep)
		os.Exit(1)
	}

	CreateArchive(root, collection, archiveParams{
		exclude:     allCliFlags.excludeDirs,
		interval:    allCliFlags.sweepInterval,
		maxProc:     allCliFlags.maxProc,
		dryRun:      allCliFlags.dryRun,
		deleteLocal: allCliFlags.deleteLocal,
	})
}

func CreateArchive(root string, archiveRoot string, params archiveParams) {
	log := logs.GetLogger()

	cancelCtx, cancel := context.WithCancel(context.Background())
	setupSignalHandler(cancel)

	userPruneFn, err := valet.MakeGlobPruneFunc(params.exclude)
	if err != nil {
		log.Error().Err(err).Msg("error in default exclusion patterns")
		os.Exit(1)
	}

	defaultPruneFn, err := valet.MakeDefaultPruneFunc(root)
	if err != nil {
		log.Error().Err(err).Msg("error in exclusion patterns")
		os.Exit(1)
	}

	clientPool := ex.NewClientPool(maxClients, time.Second*1, "--silent")

	var workPlan valet.WorkPlan
	if params.dryRun {
		workPlan = valet.DryRunWorkPlan()
	} else {
		workPlan = valet.ArchiveFilesWorkPlan(root, archiveRoot, clientPool,
			params.deleteLocal)
	}

	valet.ProcessFiles(cancelCtx, valet.ProcessParams{
		Root:          root,
		MatchFunc:     valet.RequiresArchiving,
		PruneFunc:     valet.Or(userPruneFn, defaultPruneFn),
		Plan:          workPlan,
		SweepInterval: params.interval,
		MaxProc:       params.maxProc,
	})
}
