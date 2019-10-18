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

	ex "github.com/kjsanger/extendo"
	logs "github.com/kjsanger/logshim"

	"github.com/kjsanger/valet/utilities"
)

type FilePredicate func(path FilePath) (bool, error)

const Fast5Suffix string = "fast5" // The recognised suffix for fast5 files
const FastqSuffix string = "fastq" // The recognised suffix for fastq files
const CSVSuffix string = "csv"
const MarkdownSuffix string = "md"
const TxtSuffix string = "txt"
const PDFSuffix string = "pdf"
const MD5Suffix string = "md5"        // The recognised suffix for MD5 checksum files
const CompressionSuffix string = "gz" // The recognised suffix for compressed files
const LargeSize int64 = 524288000

var fast5Regex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", Fast5Suffix))
var fastqRegex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", FastqSuffix))
var txtRegex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", TxtSuffix))
var markdownRegex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", MarkdownSuffix))
var pdfRegex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", PDFSuffix))
var csvRegex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", CSVSuffix))
var compressedRegex = regexp.MustCompile(fmt.Sprintf(".*[.]%s$", CompressionSuffix))

// Matches the run ID of MinKNOW c. August 2019 for GridION and PromethION
// i.e. of the form:
//
// 20190701_1522_GA10000_FAK83493_3bba1763
//
var MinKNOWRunIDRegex = regexp.MustCompile(`^\d+_\d+_\S+_[A-Za-z0-9]+_[A-Za-z0-9]+$`)

var RequiresArchiving = Or(IsFast5, IsFastq, IsTxt, IsMarkdown, IsPDF, IsCSV)

// RequiresChecksum returns true if the argument is a regular file that is
// recognised as a checksum target and either has no checksum file, or has a
// checksum file that is stale.
var RequiresChecksum = And(
	IsRegular,
	RequiresArchiving,
	Or(Not(HasChecksumFile), HasStaleChecksumFile))

// RequiresCompression returns true if the argument is a regular file over 500MB
// that is recognised as an archive target and has no compressed version.
var RequiresCompression = And(
	IsRegular,
	IsLarge,
	RequiresArchiving,
	Or(Not(HasCompressedVersion), HasStaleCompressedFile))

var HasValidChecksumFile = Not(HasStaleChecksumFile)

// IsTrue always returns true.
func IsTrue(path FilePath) (bool, error) {
	return true, nil
}

// IsFalse always returns false
func IsFalse(path FilePath) (bool, error) {
	return false, nil
}

// statPath stats the given path if path.Info is nil.
func statPath(path FilePath) error {
	if path.Info == nil {
		info, err := os.Stat(path.Location)
		if err != nil {
			return err
		}
		path.Info = info
	}
	return nil
}

// IsDir returns true if the argument is a directory (by os.Stat).
func IsDir(path FilePath) (bool, error) {
	if err := statPath(path); err != nil {
		return false, err
	}
	return path.Info.IsDir(), nil
}

// IsRegular returns true if the argument is a regular file (by os.Stat).
func IsRegular(path FilePath) (bool, error) {
	if err := statPath(path); err != nil {
		return false, err
	}
	return path.Info.Mode().IsRegular(), nil
}

// IsLarge returns true if the argument is a larger than 500MB (by os.Stat).
func IsLarge(path FilePath) (bool, error) {
	if err := statPath(path); err != nil {
		return false, err
	}
	return path.Info.Size() > LargeSize, nil
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

// IsFast5 returns true if path matches the recognised fast5 pattern.
func IsFast5(path FilePath) (bool, error) {
	return fast5Regex.MatchString(path.Location), nil
}

// IsFastq returns true if path matches the recognised fastq pattern.
func IsFastq(path FilePath) (bool, error) {
	return fastqRegex.MatchString(path.Location), nil
}

// IsTxt returns true if path matches the recognised text file pattern.
func IsTxt(path FilePath) (bool, error) {
	return txtRegex.MatchString(path.Location), nil
}

//  IsMarkdown returns true if path matches the recognised markdown file
//  pattern.
func IsMarkdown(path FilePath) (bool, error) {
	return markdownRegex.MatchString(path.Location), nil
}

// IsPDF returns true if path matches the recognised PDF file pattern.
func IsPDF(path FilePath) (bool, error) {
	return pdfRegex.MatchString(path.Location), nil
}

// IsCSV returns true if path matches the recognised CSV file pattern.
func IsCSV(path FilePath) (bool, error) {
	return csvRegex.MatchString(path.Location), nil
}

// IsCompressed returns true if path matches the recognised compressed file
// pattern.
func IsCompressed(path FilePath) (bool, error) {
	return compressedRegex.MatchString(path.Location), nil
}

// HasChecksumFile returns true if the argument or its compressed version has a
// corresponding checksum file.
func HasChecksumFile(path FilePath) (bool, error) {
	_, err := os.Stat(ChecksumFilename(path))
	if err == nil {
		return true, err
	}

	compressed, err2 := IsCompressed(path)
	if err2 != nil {
		return false, err2
	}
	if !compressed {
		compressedPath := CompressedFilename(path)
		compressedFP, _ := MaybeFilePath(compressedPath)
		md5Path := ChecksumFilename(compressedFP)
		_, err = os.Stat(md5Path)
		if err == nil {
			return true, err
		} else if os.IsNotExist(err) {
			return false, nil
		}
	} else if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func staleSecondaryFile(path FilePath, secondary string, kind string) (bool, error) {
	secondaryInfo, err := os.Stat(secondary)
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

	if path.Info.ModTime().After(secondaryInfo.ModTime()) {
		logs.GetLogger().Debug().
			Str("path", path.Location).
			Time("data_time", path.Info.ModTime()).
			Time(kind+"_time", secondaryInfo.ModTime()).Msg("stale " + kind)
		return true, nil
	}

	return false, nil
}

// HasStaleChecksumFile returns true if the argument has a checksum file with a
// timestamp older than the argument file i.e. the argument file appears to
// have been modified since the checksum file was last modified.
//
// If the argument path does not exist, or has no checksum file, this function
// returns false.
func HasStaleChecksumFile(path FilePath) (bool, error) {
	return staleSecondaryFile(path, ChecksumFilename(path), "checksum")
}

// HasCompressedVersion returns true if the argument has a corresponding
// compressed version of the file.
func HasCompressedVersion(path FilePath) (bool, error) {
	_, err := os.Stat(CompressedFilename(path))
	if err == nil {
		return true, err
	} else if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// HasStaleCompressedFile returns true if the argument has a compressed version
// with a timestamp older than the argument file i.e. the argument file appears
// to have been modified since the compressed file was last modified.
//
// If the argument path does not exist, or has no compressed version, this
// function returns false.
func HasStaleCompressedFile(path FilePath) (bool, error) {
	return staleSecondaryFile(path, CompressedFilename(path), CompressionSuffix)
}

// IsMinKNOWRunID returns true if name is in the form of a MinNKOW run
// identifier (matches MinKNOWRunIDRegex).
func IsMinKNOWRunID(name string) bool {
	return MinKNOWRunIDRegex.MatchString(name)
}

// IsMinKNOWRunDir returns true if path is a MinKNOW run directory. This type
// of directory is located two levels down from the data directory, within an
// experiment and a sample directory and its name is a MinKNOW run identifier.
func IsMinKNOWRunDir(path FilePath) (bool, error) {
	return IsMinKNOWRunID(filepath.Base(path.Location)), nil
}

// maybeCompressedPath returns the given path if it has no compressed version,
// otherwise it returns the path of the compressed version.
func maybeCompressedPath(path FilePath) (FilePath, error) {
	compressed, err := HasCompressedVersion(path)
	if err != nil {
		return path, err
	}
	if compressed {
		path, err = NewFilePath(CompressedFilename(path))
	}
	return path, err
}

// MakeIsArchived returns a predicate that will return true if its argument has
// been successfully archived from localBase to remoteBase.
//
// The criteria for archived state are:
//
// 1. The file has a valid checksum file (not stale), otherwise there could
//    be no way to test the checksum against the checksum in the archive.
//
// 2. The data object exists in the archive.
//
// 3. The checksum of the data object in the archive matches the expected
//    checksum.
//
// 4. If maybeCompressed is true and and a compressed version exists, the file
//    being considered in 1-3 is the compressed version of the incoming file.
func MakeIsArchived(localBase string, remoteBase string, maybeCompressed bool,
	cPool *ex.ClientPool) FilePredicate {

	return func(path FilePath) (ok bool, err error) {
		if maybeCompressed {
			path, err = maybeCompressedPath(path)
			if err != nil {
				return
			}
		}

		var dest string
		dest, err = translatePath(localBase, remoteBase, path)
		if err != nil {
			return
		}

		client, err := cPool.Get()
		if err != nil {
			return
		}

		defer func() {
			err = utilities.CombineErrors(err, cPool.Return(client))
		}()

		var chkFile FilePath
		chkFile, err = NewFilePath(ChecksumFilename(path))
		if err != nil {
			return
		}

		log := logs.GetLogger()
		ok, err = HasValidChecksumFile(path)
		if err != nil || !ok {
			log.Debug().Str("path", path.Location).
				Msg("not archived: no valid checksum file")
			return
		}

		var checksum []byte
		checksum, err = ReadMD5ChecksumFile(chkFile)
		if err != nil {
			return
		}

		obj := ex.NewDataObject(client, dest)
		ok, err = obj.Exists()
		if err != nil || !ok {
			log.Debug().Str("path", path.Location).
				Msg("not archived: data object not confirmed")
			return
		}

		chk := string(checksum)
		ok, err = obj.HasValidChecksum(chk)
		if err != nil {
			return
		}

		if !ok {
			log.Debug().Str("path", path.Location).
				Str("expected_checksum", chk).
				Str("checksum", obj.Checksum()).
				Msg("not archived: checksum not confirmed")
		}

		ok, err = obj.HasValidChecksumMetadata(chk)
		if err != nil || !ok {
			log.Debug().Str("path", path.Location).
				Msg("not archived: checksum metadata not confirmed")
			return
		}

		return
	}
}

// ChecksumFilename returns the expected path of the checksum file
// corresponding to the argument
func ChecksumFilename(path FilePath) string {
	return fmt.Sprintf("%s.%s", path.Location, MD5Suffix)
}

// CompressedFilename returns the expected path of the compressed file
// corresponding to the argument
func CompressedFilename(path FilePath) string {
	return fmt.Sprintf("%s.%s", path.Location, CompressionSuffix)
}
