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
 * @file archive.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package cmd

import (
	"os"

	logs "github.com/kjsanger/logshim"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive files under a root directory",
	Long: `
valet archive provides commands to archive data files by copying them to a
remote data store, adding metadata, validating the copy and then deleting the
original file from the local disk.
`,
	Run: runArchiveCmd,
}

func init() {
	valetCmd.AddCommand(archiveCmd)
}

func runArchiveCmd(cmd *cobra.Command, args []string) {
	if err := cmd.Help(); err != nil {
		logs.GetLogger().Error().Err(err).Msg("help command failed")
		os.Exit(1)
	}
}
