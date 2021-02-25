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

	logs "github.com/wtsi-npg/logshim"
)

// Directory names within the root MinKNOW data directory (typically /data)
// that we will ignore by default.
var MinKNOWIgnore = []string{
	"epi2me_inside",
	"intermediate",
	"npg",
	"pings",
	"queued_reads",
	"reads",
	"reports",
}

// DefaultIgnorePatterns returns glob patterns matching directories in the root
// MinKNOW data directory that will be ignored by default.
func DefaultIgnorePatterns(dataDir string) ([]string, error) {
	absData, err := filepath.Abs(dataDir)
	if err != nil {
		return []string{}, err
	}

	var patterns []string
	for _, name := range MinKNOWIgnore {
		patterns = append(patterns, filepath.Join(absData, name))
	}

	return patterns, nil
}

// MakeDefaultPruneFunc returns a directory pruning function for MinKNOW data
// directory dataDir. This will exclude directories matching
// DefaultIgnorePatterns.
func MakeDefaultPruneFunc(dataDir string) (FilePredicate, error) {
	defaults, err := DefaultIgnorePatterns(dataDir)
	if err != nil {
		return nil, err
	}

	return MakeGlobPruneFunc(defaults)
}

/*
// MakeRegexPruneFunc returns a FilePredicate that will return false for any
// directory matching at least one of the regex pattern arguments. The returned
// function is intended for use as a pruning function argument to the
// valet.WatchFiles and valet.FindFiles functions.
func MakeRegexPruneFunc(patterns []string) (FilePredicate, error) {
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

	return func(fp FilePath) (bool, error) {
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

// MakeGlobPruneFunc returns a FilePredicate that will return false for any
// directory matching at least one of the glob pattern arguments. The returned
// function is intended for use as a pruning function argument to the
// valet.WatchFiles and valet.FindFiles functions.
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
