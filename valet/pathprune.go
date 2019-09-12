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
 * @file pathprune.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"path/filepath"

	logs "github.com/kjsanger/logshim"
)

/*
func MakeRegexPruneFunc(patterns []string) (valet.FilePredicate, error) {
	log := logs.GetLogger()

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

func MakeGlobPruneFunc(patterns []string) (FilePredicate, error) {
	log := logs.GetLogger()

	for _, pattern := range patterns {
		if _, err := filepath.Match(pattern, "."); err != nil {
			return nil, err
		}
	}

	return func(fp FilePath) (bool, error) {
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
				return true, filepath.SkipDir // return SkipDir to prune here
			}
		}
		return false, nil
	}, nil
}
