/*
 * Copyright (C) 2019, 2020, 2021, 2022. Genome Research Ltd. All rights
 * reserved.
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
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	ex "github.com/wtsi-npg/extendo/v2"
	logs "github.com/wtsi-npg/logshim"

	"github.com/wtsi-npg/valet/utilities"
	"github.com/wtsi-npg/valet/valet"
)

type archiveParams struct {
	deleteLocal   bool
	dryRun        bool
	exclude       []string
	sweepInterval time.Duration
	maxProc       int
	cleanupDelay  time.Duration
}

var archCreateFlags = &dataDirCliFlags{}

var archiveCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an archive of files under a root directory",
	Long: `
valet archive create will monitor a directory hierarchy and locate data files
within it that are not currently within a remote data store. valet will then
archive them and if successful, delete the archived file from the local disk.

N.B. TMPDIR use

- TMPDIR must be set to the same filesystem as the data root.

- valet will automatically exclude TMPDIR from its operations.

TMPDIR is used by valet to compress files, after which they are moved into
position using a rename operation. Renaming files will fail if attempted across
filesystem boundaries.

- Archiving files
  
  - Directory hierarchy styles supported

    - Any
  
  - File patterns supported

    - All supported for archiving

  - Checksum file patterns supported

    - (data file name).md5
`,
	Example: `
valet archive create --root /data --exclude /data/custom \
    --archive-root /seq/ont/gridion/gxb02004
    --interval 20m --verbose`,
	Run: runArchiveCreateCmd,
}

func init() {
	archiveCreateCmd.Flags().StringVarP(&archCreateFlags.localRoot,
		"root", "r", "",
		"the root directory of the monitor")

	err := archiveCreateCmd.MarkFlagRequired("root")
	if err != nil {
		logs.GetLogger().Error().
			Err(err).Msg("failed to mark --root required")
		os.Exit(1)
	}

	archiveCreateCmd.Flags().StringVarP(&archCreateFlags.archiveRoot,
		"archive-root", "a", "",
		"the archive root collection")

	err = archiveCreateCmd.MarkFlagRequired("archive-root")
	if err != nil {
		logs.GetLogger().Error().
			Err(err).Msg("failed to mark --archive-root required")
		os.Exit(1)
	}

	archiveCreateCmd.Flags().DurationVarP(&archCreateFlags.sweepInterval,
		"interval", "i", valet.DefaultSweepInterval,
		fmt.Sprintf("directory sweep interval, minimum %s",
			valet.MinSweepInterval))

	archiveCreateCmd.Flags().BoolVar(&baseFlags.dryRun,
		"dry-run", false,
		"dry-run (make no changes)")

	archiveCreateCmd.Flags().StringArrayVar(&archCreateFlags.excludeDirs,
		"exclude", []string{},
		"patterns matching directories to prune "+
			"from both monitoring and interval sweeps")

	archiveCreateCmd.Flags().BoolVar(&archCreateFlags.deleteLocal,
		"delete-on-archive", false,
		"delete local files on successful archiving")

	archiveCreateCmd.Flags().DurationVar(&archCreateFlags.cleanupDelay,
		"cleanup", valet.DefaultCleanupDelay,
		fmt.Sprintf("run directory cleanup delay, minimum %s",
			valet.MinCleanupDelay))

	archiveCmd.AddCommand(archiveCreateCmd)
}

func runArchiveCreateCmd(cmd *cobra.Command, args []string) {
	log := setupLogger(baseFlags)

	if archCreateFlags.sweepInterval < valet.MinSweepInterval {
		log.Error().Msgf("invalid sweep interval %s (must be > %s)",
			archCreateFlags.sweepInterval, valet.MinSweepInterval)
		os.Exit(1)
	}

	if archCreateFlags.cleanupDelay < valet.MinCleanupDelay {
		log.Error().Msgf("invalid cleanup delay %s (must be > %s)",
			archCreateFlags.cleanupDelay, valet.MinCleanupDelay)
		os.Exit(1)
	}

	err := CreateArchive(
		archCreateFlags.localRoot,
		archCreateFlags.archiveRoot,
		archiveParams{
			dryRun:        baseFlags.dryRun,
			maxProc:       baseFlags.maxProc,
			exclude:       archiveExcludeDirs(archCreateFlags.localRoot, archCreateFlags),
			sweepInterval: archCreateFlags.sweepInterval,
			deleteLocal:   archCreateFlags.deleteLocal,
			cleanupDelay:  archCreateFlags.cleanupDelay,
		})

	if err != nil {
		log.Error().Err(err).Msg("archive creation failed")
		os.Exit(1)
	}
}

// CreateArchive archives files found locally under root to remote archiveRoot,
// preserving the relative directory hierarchy.
func CreateArchive(root string, archiveRoot string, params archiveParams) error {
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

	userCleanupFn := valet.MakeRequiresRemoval(params.cleanupDelay)

	poolParams := ex.DefaultClientPoolParams
	clientPool := ex.NewClientPool(poolParams, "--silent")

	var workPlan valet.WorkPlan
	if params.dryRun {
		workPlan = valet.DryRunWorkPlan()
	} else {
		workPlan = valet.ArchiveFilesWorkPlan(root, archiveRoot, clientPool,
			params.deleteLocal, params.cleanupDelay)
	}

	return valet.ProcessFiles(cancelCtx, valet.ProcessParams{
		Root: root,
		MatchFunc: valet.Or(valet.RequiresCompression,
			valet.RequiresCopying, userCleanupFn),
		PruneFunc:     valet.Or(userPruneFn, defaultPruneFn),
		Plan:          workPlan,
		SweepInterval: params.sweepInterval,
		MaxProc:       params.maxProc,
	})
}

// Exclude TMPDIR if it has been set to be under the data root by the user
func archiveExcludeDirs(root string, flags *dataDirCliFlags) []string {
	tempDir := os.TempDir()
	rootContainsTemp, err := utilities.IsDescendantPath(root, tempDir)

	log := logs.GetLogger()
	if err != nil {
		log.Error().Err(err).
			Msgf("error excluding temp directory '%s' from archiving",
				os.TempDir())
		os.Exit(1)
	}

	excludeDirs := flags.excludeDirs
	if rootContainsTemp {
		log.Info().Str("root", root).Str("temp_dir", tempDir).
			Msg("excluding temp directory from the archiving process")
		excludeDirs = append(excludeDirs, tempDir)
	}
	return excludeDirs
}
