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
	"regexp"

	logf "valet/log/logfacade"
)

type FilePredicate func(path FilePath) (bool, error)

const Fast5Suffix string = "fast5" // The recognised suffix for fast5 files
const FastqSuffix string = "fastq" // The recognised suffix for fastq files
const MD5Suffix string = "md5"     // The recognised suffix for MD5 checksum files

var fast5Regex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", Fast5Suffix))
var fastqRegex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", FastqSuffix))

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

// IsDir returns true if the argument is a directory.
func IsDir(path FilePath) (bool, error) {
	return path.Info.IsDir(), nil
}

// IsRegular returns true if the argument is a regular file (by os.Stat).
func IsRegular(path FilePath) (bool, error) {
	return path.Info.Mode().IsRegular(), nil
}

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
		logf.GetLogger().Debug().
			Str("path", path.Location).
			Time("data_time", path.Info.ModTime()).
			Time("checksum_time", chkInfo.ModTime()).Msg("stale checksum")
		return true, nil
	}

	return false, nil
}
