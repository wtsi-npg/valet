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
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	logf "valet/log/logfacade"
	"valet/valet"
)

const defaultSweep = 5 * time.Minute
const minSweep = 30 * time.Second

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
	setupLogger(allCliFlags)
	valetCmd.AddCommand(checksumCmd)
}

func runChecksumCmd(cmd *cobra.Command, args []string) {
	if err := cmd.Help(); err != nil {
		logf.GetLogger().Error().Err(err).Msg("help command failed")
		os.Exit(1)
	}
}

/*
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
*/

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

func mergeFileChannels(
	x <-chan valet.FilePath,
	y <-chan valet.FilePath) chan valet.FilePath {
	merged := make(chan valet.FilePath)

	log := logf.GetLogger()

	go func() {
		defer close(merged)

		for x != nil || y != nil {
			select {
			case p, ok := <-x:
				if ok {
					log.Debug().Msg("merging an x path")
					merged <- p
				} else {
					log.Debug().Msg("x was closed")
					x = nil
				}

			case p, ok := <-y:
				if ok {
					log.Debug().Msg("merging a y path")
					merged <- p
				} else {
					log.Debug().Msg("y was closed")
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

		for x != nil || y != nil {
			select {
			case p, ok := <-x:
				if ok {
					merged <- p
				} else {
					x = nil
				}

			case p, ok := <-y:
				if ok {
					merged <- p
				} else {
					y = nil
				}
			}
		}
	}()

	return merged
}
