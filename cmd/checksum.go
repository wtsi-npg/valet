/*
 * Copyright (C) 2019, 2020. Genome Research Ltd. All rights reserved.
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
	"os"

	"github.com/spf13/cobra"
	logs "github.com/wtsi-npg/logshim"
)

var checksumFlags = &dataDirCliFlags{}

var checksumCmd = &cobra.Command{
	Use:   "checksum",
	Short: "Manage checksum files under a root directory",
	Long: `
valet checksum provides commands to manage checksum files that accompany data
files.
`,
	Run: runChecksumCmd,
}

func init() {
	valetCmd.AddCommand(checksumCmd)
}

func runChecksumCmd(cmd *cobra.Command, args []string) {
	if err := cmd.Help(); err != nil {
		logs.GetLogger().Error().Err(err).Msg("help command failed")
		os.Exit(1)
	}
}
