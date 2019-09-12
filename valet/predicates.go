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
 * @file predicates.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	logs "github.com/kjsanger/logshim"
)

type FilePredicate func(path FilePath) (bool, error)

const Fast5Suffix string = "fast5" // The recognised suffix for fast5 files
const FastqSuffix string = "fastq" // The recognised suffix for fastq files
const MD5Suffix string = "md5"     // The recognised suffix for MD5 checksum files

var fast5Regex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", Fast5Suffix))
var fastqRegex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", FastqSuffix))

// Matches the run ID of MinKNOW c. August 2019 for GridION and PromethION
// i.e. of the form:
//
// 20190701_1522_GA10000_FAK83493_3bba1763
//
var MinKNOWRunIDRegex = regexp.MustCompile(`\d+_\d+_GA\d+_F[A-Za-z0-9]+_[A-Za-z0-9]+`)

// RequiresChecksum returns true if the argument is a regular file that is
// recognised as a checksum target and either has no checksum file, or has a
// checksum file that is stale.
var RequiresChecksum = And(
	IsRegular,
	Or(IsFast5Match, IsFastqMatch),
	Or(Not(HasChecksumFile), HasStaleChecksumFile))

// IsTrue always returns true.
func IsTrue(path FilePath) (bool, error) {
	return true, nil
}

// IsFalse always returns false
func IsFalse(path FilePath) (bool, error) {
	return false, nil
}

// IsDir returns true if the argument is a directory (by os.Stat).
func IsDir(path FilePath) (bool, error) {
	if path.Info == nil {
		info, err := os.Stat(path.Location)
		if err != nil {
			return false, err
		}
		path.Info = info
	}
	return path.Info.IsDir(), nil
}

// IsRegular returns true if the argument is a regular file (by os.Stat).
func IsRegular(path FilePath) (bool, error) {
	if path.Info == nil {
		info, err := os.Stat(path.Location)
		if err != nil {
			return false, err
		}
		path.Info = info
	}
	return path.Info.Mode().IsRegular(), nil
}

// And returns a predicate that returns true if all its arguments return true,
// or returns false otherwise.
func And(predicates ...FilePredicate) FilePredicate {
	return func(path FilePath) (bool, error) {
		for _, p := range predicates {
			val, err := p(path)
			if err != nil {
				return false, err
			} else if !val {
				return false, nil
			}
		}
		return true, nil
	}
}

// And returns a predicate that returns true if any of its arguments return
// true, or returns false otherwise.
func Or(predicates ...FilePredicate) FilePredicate {
	return func(path FilePath) (bool, error) {
		for _, p := range predicates {
			val, err := p(path)
			if err != nil {
				return false, err
			} else if val {
				return true, nil
			}
		}
		return false, nil
	}
}

// Not returns a predicate that returns true if its argument returns false, or
// returns false otherwise.
func Not(predicate FilePredicate) FilePredicate {
	return func(path FilePath) (bool, error) {
		val, err := predicate(path)
		if err != nil {
			return false, err
		} else if val {
			return false, nil
		}
		return true, nil
	}
}

// IsFast5Match returns true if path matches the recognised fast5 pattern.
func IsFast5Match(path FilePath) (bool, error) {
	return fast5Regex.MatchString(path.Location), nil
}

// IsFastqMatch returns true if path matches the recognised fastq pattern.
func IsFastqMatch(path FilePath) (bool, error) {
	return fastqRegex.MatchString(path.Location), nil
}

// HasChecksumFile returns true if the argument has a corresponding checksum
// file.
func HasChecksumFile(path FilePath) (bool, error) {
	_, err := os.Stat(ChecksumFilename(path))
	if err == nil {
		return true, err
	} else if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// HasStaleChecksumFile returns true if the argument has a checksum file with a
// timestamp older than the argument file i.e. the argument file appears to
// have been modified since the checksum file was last modified.
//
// If the argument path does not exist, or has no checksum file, this function
// returns false.
func HasStaleChecksumFile(path FilePath) (bool, error) {
	chkInfo, err := os.Stat(ChecksumFilename(path))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	if path.Info == nil {
		path.Info, err = os.Stat(path.Location)
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if path.Info.ModTime().After(chkInfo.ModTime()) {
		logs.GetLogger().Debug().
			Str("path", path.Location).
			Time("data_time", path.Info.ModTime()).
			Time("checksum_time", chkInfo.ModTime()).Msg("stale checksum")
		return true, nil
	}

	return false, nil
}

// IsMinKNOWRunDir returns true if path is a MinKNOW run directory. This type
// of directory is located two levels down from the data directory, within an
// experiment and a sample directory.
func IsMinKNOWRunDir(path FilePath) (bool, error) {
	return MinKNOWRunIDRegex.MatchString(filepath.Base(path.Location)), nil
}

// MakeMinKNOWExptDirPred returns a predicate that tests a directory to see
// whether it is a MinKNOW run directory.
//
// A run directory is defined as:
//
// - a directory
// - that is directly contained in the MinNKOW data directory (usually /data)
// - that contains one or more directories (sample directories) that themselves
//   contain MinKNOW run directories.
// - where a MinKNOW run directory is defined by the directory name matching
//   MinKNOWRunIDRegex.
//
// e.g. /data/27 is an experiment directory
//
// /data/27/ABCD123456/20190701_1522_GA10000_FAK83493_3bba1763
//
// This function is useful because the MinKNOW data directory may contain
// directories other than those containing sequencing results.
func MakeMinKNOWExptDirPred(dataDir string) (FilePredicate, error) {
	root, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}

	runPattern :=  fmt.Sprintf("%s/*/*", filepath.Clean(root))

	pred := func(path FilePath) (bool, error) {
		var err error

		if path.Info == nil {
			path.Info, err = os.Stat(path.Location)
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		if !path.Info.IsDir() {
			return false, nil
		}

		contents, err := filepath.Glob(runPattern)
		for _, path := range contents {
			if MinKNOWRunIDRegex.MatchString(path) {
				return true, err
			}
		}

		return false, err
	}

	return pred, err
}

// ChecksumFilename returns the expected path of the checksum file
// corresponding to the argument
func ChecksumFilename(path FilePath) string {
	return fmt.Sprintf("%s.%s", path.Location, MD5Suffix)
}
