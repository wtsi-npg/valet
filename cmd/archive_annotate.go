/*
 * Copyright (C) 2020. Genome Research Ltd. All rights reserved.
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
 * @file archive_annotate.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package cmd

import (
	"os"

	ex "github.com/kjsanger/extendo/v2"
	logs "github.com/kjsanger/logshim"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kjsanger/valet/utilities"
	"github.com/kjsanger/valet/valet"
)

var archAnnotateFlags = &dataFileCliFlags{}

var archiveAnnotateCmd = &cobra.Command{
	Use:   "annotate",
	Short: "Annotate archived data",
	Long: `
valet annotate ont will use metadata from a local run folder to annotate the
corresponding run data within a remote data store.
`,
	Example: `
valet annotate ont \ 
  --path /data/66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f \
  --archive-path /archive/66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f \
  --verbose
`,
	Run: runArchiveAnnotateCmd,
}

func init() {
	archiveAnnotateCmd.Flags().StringVarP(&archAnnotateFlags.localPath,
		"path", "p", "",
		"the local path of the annotation file")

	err := archiveAnnotateCmd.MarkFlagRequired("path")
	if err != nil {
		logs.GetLogger().Error().
			Err(err).Msg("failed to mark --path required")
		os.Exit(1)
	}

	archiveAnnotateCmd.Flags().StringVarP(&archAnnotateFlags.archivePath,
		"archive-path", "a", "",
		"the archive path of the annotation file")

	err = archiveAnnotateCmd.MarkFlagRequired("archive-path")
	if err != nil {
		logs.GetLogger().Error().
			Err(err).Msg("failed to mark --archive-path required")
		os.Exit(1)
	}

	archiveCmd.AddCommand(archiveAnnotateCmd)
}

func runArchiveAnnotateCmd(cmd *cobra.Command, args []string) {
	log := setupLogger(baseFlags)

	err := AnnotateArchive(archAnnotateFlags.localPath,
		archAnnotateFlags.archivePath)
	if err != nil {
		log.Error().Err(err).Msg("archive annotation failed")
		os.Exit(1)
	}

	log.Info().Str("path", archAnnotateFlags.localPath).
		Str("to", archAnnotateFlags.archivePath).
		Msg("annotation confirmed")
}

// AnnotateArchive creates or updates any remote annotation originating from
// a file at localPath which is archived at archivePath.
func AnnotateArchive(localPath string, archivePath string) (err error) { // NRV
	var fp valet.FilePath
	fp, err = valet.NewFilePath(localPath)
	if err != nil {
		return
	}

	var ok bool
	if ok, err = valet.IsMinKNOWReport(fp); err != nil {
		return
	}
	if !ok {
		return errors.Errorf("'%s' does not appear to be a MinKNOW " +
			"report file", localPath)
	}

	var report valet.MinKNOWReport
	if report, err = valet.ParseMinKNOWReport(localPath); err != nil {
		return
	}

	cPool := ex.NewClientPool(ex.DefaultClientPoolParams, "--silent")

	var client *ex.Client
	if client, err = cPool.Get(); err != nil {
		return
	}
	defer func() {
		err = utilities.CombineErrors(err, cPool.Return(client))
	}()

	obj := ex.NewDataObject(client, archivePath)
	if err = valet.AddMinKNOWReportAnnotation(obj, report); err != nil {
		return
	}

	if ok, err = valet.HasValidReportAnnotation(obj, report); err != nil {
		return
	}
	if !ok {
		return errors.Errorf("metadata from MinNOW report file '%s' " +
			"was not confirmed for '%s'", localPath, archivePath)
	}

	return nil
}
