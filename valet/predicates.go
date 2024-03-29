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
 * @file predicates.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pkg/errors"
	ex "github.com/wtsi-npg/extendo/v2"
	logs "github.com/wtsi-npg/logshim"

	"github.com/wtsi-npg/valet/utilities"
)

type FilePredicate func(path FilePath) (bool, error)

// We match data files explicitly because various versions of MinKNOW have put
// temporary files into the data area. If these were whisked off to iRODS, it
// would at best put junk into the archive and at worst break the run processing
// on the instrument.

const Fast5Suffix string = "fast5"
const FastqSuffix string = "fastq"
const BAISuffix string = "bai"
const BAMSuffix string = "bam"
const BEDSuffix string = "bed"
const CSVSuffix string = "csv"
const HTMLSuffix string = "html"
const MarkdownSuffix string = "md"
const JSONSuffix string = "json"
const TxtSuffix string = "txt"
const TSVSuffix string = "tsv"
const PDFSuffix string = "pdf"
const POD5Suffix string = "pod5"
const MD5Suffix string = "md5" // The recognised suffix for MD5 checksum files
const GzipSuffix string = "gz"

var fast5Regex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", Fast5Suffix))
var fastqRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", FastqSuffix))
var baiRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", BAISuffix))
var bamRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", BAMSuffix))
var bedRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", BEDSuffix))
var htmlRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", HTMLSuffix))
var jsonRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", JSONSuffix))
var txtRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", TxtSuffix))
var tsvRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", TSVSuffix))
var markdownRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", MarkdownSuffix))
var pdfRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", PDFSuffix))
var pod5Regex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", POD5Suffix))
var csvRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", CSVSuffix))
var gzipRegex = regexp.MustCompile(fmt.Sprintf("(?i).*[.]%s$", GzipSuffix))
var reportRegex = regexp.MustCompile(fmt.Sprintf("(?i)report.*[.]%s$", MarkdownSuffix))

// IsFast5 returns true if path matches the recognised fast5 pattern.
var IsFast5 = makeNoCompFilePredicate(fast5Regex)

// IsPOD5 returns true if path matches the recognised pod5 pattern.
var IsPOD5 = makeNoCompFilePredicate(pod5Regex)

// IsFastq returns true if path matches the recognised fastq pattern. Supports
// compressed versions.
var IsFastq = makeCompFilePredicate(fastqRegex)

// IsBED returns true if path matches the recognised BED file pattern.
// Supports compressed versions.
var IsBED = makeCompFilePredicate(bedRegex)

// IsBAI returns true if path matches the recognised BAI file pattern.
var IsBAI = makeNoCompFilePredicate(baiRegex)

// IsBAM returns true if path matches the recognised BAM file pattern.
var IsBAM = makeNoCompFilePredicate(bamRegex)

// IsTxt returns true if path matches the recognised text file pattern.
// Supports compressed versions.
var IsTxt = makeCompFilePredicate(txtRegex)

// IsMarkdown returns true if path matches the recognised markdown file
// pattern. Supports compressed versions.
var IsMarkdown = makeCompFilePredicate(markdownRegex)

// IsPDF returns true if path matches the recognised PDF file pattern.
var IsPDF = makeNoCompFilePredicate(pdfRegex)

// IsHTML returns true if path matches the recognised HTML file pattern.
// Supports compressed versions.
var IsHTML = makeCompFilePredicate(htmlRegex)

// IsCSV returns true if path matches the recognised CSV file pattern.
// Supports compressed versions.
var IsCSV = makeCompFilePredicate(csvRegex)

// IsTSV returns true if path matches the recognised TSV file pattern.
// Supports compressed versions.
var IsTSV = makeCompFilePredicate(tsvRegex)

// IsJSON returns true if path matches the recognised JSON file pattern.
// Supports compressed versions.
var IsJSON = makeCompFilePredicate(jsonRegex)

// MinKNOWRunIDRegex matches the run ID of MinKNOW c. August 2019 for GridION
// and PromethION i.e. of the form:
//
// 20190701_1522_GA10000_FAK83493_3bba1763
//
var MinKNOWRunIDRegex = regexp.MustCompile(`^\d+_\d+_\S+_[A-Za-z0-9]+_[A-Za-z0-9]+$`)

var RequiresCopying = Or(
	And(IsBED, IsCompressed),
	And(IsCSV, IsCompressed),
	And(IsFastq, IsCompressed),
	And(IsJSON, IsCompressed),
	And(IsTxt, IsCompressed),
	IsBAI,
	IsBAM,
	IsFast5,
	IsHTML,
	IsMarkdown,
	IsPDF,
	IsPOD5,
	IsTSV,
)

// RequiresChecksum returns true if the argument is a regular file that is
// recognised as a checksum target and either has no checksum file, or has a
// checksum file that is stale.
var RequiresChecksum = And(
	IsRegular,
	RequiresCopying,
	Or(Not(HasChecksumFile), HasStaleChecksumFile))

var HasValidChecksumFile = Not(HasStaleChecksumFile)

var RequiresCompression = And(
	Or(
		IsBED,
		IsCSV,
		IsFastq,
		IsJSON,
		IsTxt,
	),
	Not(IsCompressed),
	Not(HasCompressedVersion))

var RequiresAnnotation = IsMinKNOWReport

// IsTrue always returns true.
func IsTrue(_ FilePath) (bool, error) {
	return true, nil
}

// IsFalse always returns false
func IsFalse(_ FilePath) (bool, error) {
	return false, nil
}

// IsDir returns true if the argument is a directory (by os.Stat).
func IsDir(path FilePath) (bool, error) {
	return path.Info.IsDir(), nil
}

// IsRegular returns true if the argument is a regular file (by os.Stat).
func IsRegular(path FilePath) (bool, error) {
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

// Or returns a predicate that returns true if any of its arguments return
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

// IsCompressed returns true if the path matches the recognised compressed file
// pattern (simply *.gz at the moment).
func IsCompressed(path FilePath) (bool, error) {
	return gzipRegex.MatchString(path.Location), nil
}

// HasCompressedVersion returns true if the argument is not a compressed file
// and has a corresponding compressed version.
func HasCompressedVersion(path FilePath) (bool, error) {
	compressed, err := IsCompressed(path)

	if err == nil && !compressed {
		_, err := os.Stat(path.CompressedFilename())
		if err == nil {
			logs.GetLogger().Debug().Str("path", path.Location).
				Msg("compressed version present")
			return true, err
		} else if os.IsNotExist(err) {
			return false, nil
		}
	}

	return false, err
}

// HasChecksumFile returns true if the argument has a corresponding checksum
// file.
func HasChecksumFile(path FilePath) (bool, error) {
	_, err := os.Stat(path.ChecksumFilename())
	if err == nil {
		logs.GetLogger().Debug().Str("path", path.Location).
			Msg("checksum file present")
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
	chkInfo, err := os.Stat(path.ChecksumFilename())
	if err != nil {
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

// IsMinKNOWRunID returns true if name is in the form of a MinKNOW run
// identifier (matches MinKNOWRunIDRegex).
func IsMinKNOWRunID(name string) bool {
	return MinKNOWRunIDRegex.MatchString(name)
}

// IsMinKNOWRunDir returns true if path is a MinKNOW run directory. This type
// of directory is located two levels down from the data directory, within an
// experiment and a sample directory and its name is a MinKNOW run identifier.
func IsMinKNOWRunDir(path FilePath) (bool, error) {
	if path.Info.IsDir() {
		return IsMinKNOWRunID(filepath.Base(path.Location)), nil
	}
	return false, nil
}

// IsMinKNOWReport returns true if path is a MinKNOW run report file. This
// file is Markdown that contains a section of JSON metadata describing details
// of the run.
func IsMinKNOWReport(path FilePath) (bool, error) {
	return reportRegex.MatchString(path.Location), nil
}

// MakeIsOlderThan returns a predicate that will return true if its argument is
// older than the specified duration.
func MakeIsOlderThan(duration time.Duration) FilePredicate {
	return func(path FilePath) (bool, error) {
		return time.Now().Sub(path.Info.ModTime()) > duration, nil
	}
}

// MakeRequiresRemoval returns a predicate that will return true if its argument
// is a run directory that may be removed because it is older than the specified
// duration.
func MakeRequiresRemoval(duration time.Duration) FilePredicate {
	return And(IsMinKNOWRunDir, MakeIsOlderThan(duration))
}

// MakeIsCopied returns a predicate that will return true if its argument has
// been successfully copied from localBase to remoteBase, and no errors occur
// while confirming this.
//
// The criteria for copied state are:
//
// 1. The file has a valid checksum file (not stale), otherwise there could
//    be no way to test the checksum against the checksum in the archive.
//
// 2. The data object exists in the archive.
//
// 3. The checksum of the data object in the archive matches the expected
//    checksum.
//
// 4. The data object has metadata under the "md5" key whose value matches the
//    checksum.
func MakeIsCopied(localBase string, remoteBase string,
	cPool *ex.ClientPool) FilePredicate {

	return func(path FilePath) (ok bool, err error) { // NRV
		defer func() {
			if err != nil {
				err = errors.Wrap(err, "IsCopied")
			}
		}()

		var dest string
		dest, err = translatePath(localBase, remoteBase, path)
		if err != nil {
			return false, err
		}

		client, err := cPool.Get()
		if err != nil {
			return false, err
		}

		defer func() {
			err = utilities.CombineErrors(err, cPool.Return(client))
		}()

		log := logs.GetLogger()
		obj := ex.NewDataObject(client, dest)
		ok, err = obj.Exists()
		if err != nil || !ok {
			log.Debug().Str("path", path.Location).
				Str("to", obj.RodsPath()).
				Msg("copy NOT confirmed")
			return false, err
		}

		ok, err = validateObjChecksum(path, obj)
		if !ok || err != nil {
			return ok, err
		}

		log.Debug().Str("path", path.Location).
			Str("to", obj.RodsPath()).
			Str("checksum", obj.Checksum()).
			Msg("copy confirmed")

		return true, err
	}
}

// MakeIsAnnotated returns a predicate that will return true if its argument has
// had its associated metadata annotated in iRODS, and no errors occur while
// confirming this.
//
// The criteria for annotated state are:
//
// 1. The metadata associated with the file has been obtained e.g. parsed from
//    a file.
//
// 2. The metadata are annotated in iRODS.
//
// Note that is not testing for the presence of a specific data object e.g. the
// report file that contained the metadata. That is achieved using the IsCopied
// predicate.
func MakeIsAnnotated(localBase string, remoteBase string,
	cPool *ex.ClientPool) FilePredicate {

	return func(path FilePath) (ok bool, err error) { // NRV
		defer func() {
			if err != nil {
				err = errors.Wrap(err, "IsAnnotated")
			}
		}()

		var dest string
		dest, err = translatePath(localBase, remoteBase, path)
		if err != nil {
			return false, err
		}

		var client *ex.Client
		client, err = cPool.Get()
		if err != nil {
			return false, err
		}

		defer func() {
			err = utilities.CombineErrors(err, cPool.Return(client))
		}()

		var isReport bool
		isReport, err = IsMinKNOWReport(path)
		if err != nil {
			return false, err
		}

		log := logs.GetLogger()
		if !isReport {
			log.Debug().Str("path", path.Location).
				Msg("not a MinKNOW report, annotation NOT confirmed")
			return false, err
		}

		report, err := ParseMinKNOWReport(path.Location)
		if err != nil {
			return false, err
		}

		obj := ex.NewDataObject(client, dest)
		ok, err = HasValidReportAnnotation(obj, report)
		if !ok || err != nil {
			return false, err
		}

		return true, err
	}
}

// HasValidReportAnnotation returns true if the metadata in report, which has
// been archived as obj, is up-to-date in the remote archive.
func HasValidReportAnnotation(obj *ex.DataObject, report MinKNOWReport) (bool, error) {
	log := logs.GetLogger()

	// The metadata to check is on the collection containing the file in
	// iRODS
	coll := obj.Parent()
	_, err := coll.FetchMetadata()
	if err != nil {
		return false, err
	}

	metadata, err := report.AsEnhancedMetadata()
	if err != nil {
		log.Error().Err(err).
			Str("path", coll.RodsPath()).
			Msg("report metadata invalid")
	}

	if !coll.HasAllMetadata(metadata) {
		for _, avu := range metadata {
			if !coll.HasMetadatum(avu) {
				log.Debug().Str("path", coll.RodsPath()).
					Str("attr", avu.Attr).
					Str("value", avu.Value).Msg("missing this AVU")
			}
		}

		log.Debug().Str("path", report.Path).
			Str("to", coll.RodsPath()).
			Msg("report metadata NOT confirmed")

		return false, nil
	}

	for _, avu := range metadata {
		log.Debug().Str("path", report.Path).
			Str("attribute", avu.Attr).
			Str("value", avu.Value).
			Msg("report metadata confirmed")
	}

	return true, nil
}

func makeNoCompFilePredicate(regex *regexp.Regexp) func(path FilePath) (bool, error) {
	return func(path FilePath) (bool, error) {
		return regex.MatchString(path.Location), nil
	}
}

func makeCompFilePredicate(regex *regexp.Regexp) func(path FilePath) (bool, error) {
	return func(path FilePath) (bool, error) {
		return regex.MatchString(path.UncompressedFilename()), nil
	}
}

// validateObjChecksum checks that the data file at path has a corresponding
// checksum file, that the checksum in that file is the same as that recorded
// for the corresponding data object in iRODS, and that the data object has the
// same checksum present in its metadata.
func validateObjChecksum(path FilePath, obj *ex.DataObject) (bool, error) {
	log := logs.GetLogger()

	chkFile, err := NewFilePath(path.ChecksumFilename())
	if err != nil {
		return false, err
	}

	ok, err := HasValidChecksumFile(path)
	if err != nil || !ok {
		log.Debug().Str("path", path.Location).
			Msg("valid checksum file NOT present")
		return false, err
	}

	checksum, err := ReadMD5ChecksumFile(chkFile)
	if err != nil {
		log.Debug().Str("path", path.Location).
			Msg("checksum file NOT readable")
		return false, err
	}

	chk := string(checksum)
	ok, err = obj.HasValidChecksum(chk)
	if err != nil || !ok {
		log.Debug().Str("path", path.Location).
			Str("expected_checksum", chk).
			Str("checksum", obj.Checksum()).
			Msg("checksum NOT confirmed")
		return false, err
	}

	ok, err = obj.HasValidChecksumMetadata(chk)
	if err != nil || !ok {
		log.Debug().Str("path", path.Location).
			Msg("checksum metadata NOT confirmed")
		return false, err
	}

	return true, nil
}
