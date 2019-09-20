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
 * @file workfunc.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"

	ex "github.com/kjsanger/extendo"
	"github.com/pkg/errors"

	"github.com/kjsanger/valet/utilities"

	logs "github.com/kjsanger/logshim"
)

// WorkFunc is a worker function used by ProcessFiles.
type WorkFunc func(path FilePath) error

// Work describes a function to be executed and the rank of the execution. When
// there is a choice of Work to be executed, Work with a smallest Rank value
// (i.e. highest rank) is performed first. In the case of a tie, either Work
// may be selected for execution.
type Work struct {
	WorkFunc WorkFunc // A WorkFunc to execute
	Rank     uint16   // The rank of the work
}

// WorkArr is a series of Work to be executed in ascending rank order.
type WorkArr []Work

func (s WorkArr) Len() int {
	return len(s)
}

func (s WorkArr) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s WorkArr) Less(i, j int) bool {
	return s[i].Rank < s[j].Rank
}

func (s WorkArr) IsEmpty() bool {
	return len(s) == 0
}

// WorkMatch is an association between a FilePredicate and Work to be done. If
// the predicate returns true then the work will be done.
type WorkMatch struct {
	pred FilePredicate // Predicate to match against candidate FilePath
	work Work          // Work to be executed on a matching FilePath
	desc string        // A short description of the match criteria and work
}

// WorkPlan is a slice of WorkMatches. Where more than one Work is matched,
// they will be done in rank order.
type WorkPlan []WorkMatch

func (p WorkPlan) Len() int {
	return len(p)
}

func (p WorkPlan) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p WorkPlan) Less(i, j int) bool {
	return p[i].work.Rank < p[j].work.Rank
}

func (p WorkPlan) IsEmpty() bool {
	return len(p) == 0
}

// DryRunWorkPlan matches any FilePath and does DoNothing Work.
func DryRunWorkPlan() WorkPlan {
	return []WorkMatch{{
		pred: IsTrue,
		work: Work{WorkFunc: DoNothing},
		desc: "IsTrue : DoNothing"}}
}

// CreateChecksumWorkPlan manages checksum files.
func CreateChecksumWorkPlan() WorkPlan {
	return []WorkMatch{{
		pred: RequiresChecksum,
		work: Work{WorkFunc: CreateOrUpdateMD5ChecksumFile},
		desc: "RequiresChecksum : CreateOrUpdateMD5ChecksumFile"}}
}

// ChecksumStateWorkPlan counts files that do not have a checksum.
func ChecksumStateWorkPlan(countFunc WorkFunc) WorkPlan {
	return []WorkMatch{{
		pred: RequiresChecksum,
		work: Work{WorkFunc: countFunc},
		desc: "RequiresChecksum : Count"}}
}

func ArchiveFilesWorkPlan(localBase string, remoteBase string,
	cPool *ex.ClientPool) WorkPlan {

	archiver := MakeArchiver(localBase, remoteBase, cPool)

	return []WorkMatch{
		{
			pred: RequiresChecksum,
			work: Work{WorkFunc: CreateOrUpdateMD5ChecksumFile, Rank: 1},
			desc: "RequiresChecksum : CreateOrUpdateMD5ChecksumFile",
		},
		{
			pred: RequiresArchiving,
			work: Work{WorkFunc: archiver, Rank: 2},
			desc: "RequiresArchiving : Archive",
		},
		{
			pred: MakeIsArchived(localBase, remoteBase, cPool),
			work: Work{WorkFunc: RemoveFile, Rank: 3},
		},
		{
			pred: HasChecksumFile,
			work: Work{WorkFunc:RemoveMD5ChecksumFile, Rank:4},
		},
	}
}

// DoNothing does nothing apart from log at debug level that it has been
// called. It is used to implement dry-run operations.
func DoNothing(path FilePath) error {
	logs.GetLogger().Debug().
		Str("path", path.Location).Msg("would work on this")
	return nil
}

// CreateOrUpdateMD5ChecksumFile calculates a checksum for the file at path and
// writes it to a new checksum file as a hex-encoded string. This function only
// operates when there is no existing checksum file, or when the existing
// checksum file is stale (its last modified time is older than the last
// modified time of path). If the checksum file is stale this function deletes
// it before creating a new one.
func CreateOrUpdateMD5ChecksumFile(path FilePath) error {
	log := logs.GetLogger()

	staleFile, err := HasStaleChecksumFile(path)
	if err != nil {
		log.Error().Err(err).Msg("staleFile checksum detection failed")
		return err
	}

	if staleFile {
		return UpdateMD5ChecksumFile(path)
	}

	hasFile, err := HasChecksumFile(path)
	if err != nil {
		return err
	}

	if !hasFile {
		err = CreateMD5ChecksumFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateMD5ChecksumFile calculates a checksum file for the data file at path
// with contents as a hex-encoded string. It raises an error if the checksum
// file already exists.
func CreateMD5ChecksumFile(path FilePath) error {
	md5sum, err := CalculateFileMD5(path)
	if err != nil {
		return err
	}

	return createMD5File(ChecksumFilename(path), md5sum)
}

// UpdateMD5ChecksumFile removes the existing checksum file, if it exists and
// creates a new one.
func UpdateMD5ChecksumFile(path FilePath) error {
	log := logs.GetLogger()

	if rerr := RemoveMD5ChecksumFile(path); rerr != nil {
		return rerr
	}
	log.Info().Str("path", path.Location).
		Msg("removed stale MD5 file")

	if cerr := CreateMD5ChecksumFile(path); cerr != nil {
		log.Error().Err(cerr).
			Str("path", path.Location).
			Msg("failed to create a new MD5 file")
		return cerr
	}

	return nil
}

// RemoveMD5ChecksumFile removes the MD5 checksum file corresponding to path.
// If the file does not exist by the time removal is attempted, no error is
// raised.
func RemoveMD5ChecksumFile(path FilePath) error {
	err := os.Remove(ChecksumFilename(path))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// CalculateFileMD5 returns the MD5 checksum of the file at path.
func CalculateFileMD5(path FilePath) (md5sum []byte, err error) { // NRV
	f, err := os.Open(path.Location)
	if err != nil {
		return
	}

	defer func() {
		err = utilities.CombineErrors(err, f.Close())
	}()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}
	md5sum = h.Sum(nil)
	return
}

// ReadMD5ChecksumFile reads and returns a checksum from a local file created by
// CreateMD5ChecksumFile. It trims any whitespace (including any newline) from
// the beginning and end of the checksum.
func ReadMD5ChecksumFile(path FilePath) (md5sum []byte, err error) { // NRV
	f, err := os.Open(path.Location)
	if err != nil {
		return
	}

	defer func() {
		err = utilities.CombineErrors(err, f.Close())
	}()

	md5sum, err = bufio.NewReader(f).ReadBytes('\n')
	if err != nil {
		return
	}
	md5sum = bytes.TrimSpace(md5sum)

	return
}

// MakeArchiver returns a WorkFunc capable of archiving files to iRODS. Each
// file passed to the WorkFunc will have its path relative to localBase
// calculated. This relative path will then be appended to remoteBase to give
// the full destination path in iRODS. E.g.
//
// localBase        = /a/b/c
// remoteBase       = /zone1/x/y
//
// file path        = /a/b/c/d/e/f.fast5
//
// therefore:
//
// relative path    = ./d/e/f.txt
// destination path = /zone1/x/y/d/e/f.fast5
//
// Any leading iRODS collections will be created by the WorkFunc as required.
//
// WorkFunc prerequisites: CreateOrUpdateMD5ChecksumFile
//
// i.e. files for archiving are expected to have an MD5 checksum file.
func MakeArchiver(localBase string, remoteBase string,
	cPool *ex.ClientPool) WorkFunc {
	return func(path FilePath) (err error) { // NRV
		var dst string
		dst, err = translatePath(localBase, remoteBase, path)

		var chkFile FilePath
		chkFile, err = NewFilePath(ChecksumFilename(path))
		if err != nil {
			return
		}

		var checksum []byte
		checksum, err = ReadMD5ChecksumFile(chkFile)

		log := logs.GetLogger()
		log.Info().Str("src", path.Location).Str("to", dst).
			Str("checksum", string(checksum)).Msg("archiving")

		var client *ex.Client
		client, err = cPool.Get()
		if err != nil {
			return
		}

		defer func() {
			err = utilities.CombineErrors(err, cPool.Return(client))
		}()

		coll := ex.NewCollection(client, filepath.Dir(dst))
		err = coll.Ensure()
		if err != nil {
			return
		}

		chk := string(checksum)
		if _, err = ex.ArchiveDataObject(client, path.Location, dst, chk,
			ex.MakeCreationMetadata(chk)) ; err != nil {
			return
		}

		log.Info().Str("path", path.Location).Str("to", dst).
			Str("checksum", string(checksum)).Msg("archived")
		return
	}
}

func RemoveFile(path FilePath) error {
	logs.GetLogger().Info().Str("path", path.Location).Msg("deleting")
	return os.Remove(path.Location)
}

// doWork accepts a candidate FilePath and a WorkPlan and returns Work
// encapsulating all the work in the WorkPlan. If no work is required for the
// FilePath, it returns DoNothing Work.
//
// All predicates are evaluated as any work is done, therefore if some
// predicates are true only after earlier work in the WorkPlan is complete,
// they will pass, provided work is ranked in the appropriate order.
func doWork(path FilePath, plan WorkPlan) (Work, error) {

	if plan.IsEmpty() {
		return Work{WorkFunc:DoNothing}, nil
	}

	workFunc := func(fp FilePath) error {
		wp := plan
		sort.Sort(wp)

		log := logs.GetLogger()

		for _, wm := range wp {
			ok, err := wm.pred(fp)
			if err != nil {
				return err
			}

			if ok {
				log.Debug().Str("path", fp.Location).
					Str("desc", wm.desc).
					Uint64("rank", uint64(wm.work.Rank)).
					Msg("match, doing work")

				if err := wm.work.WorkFunc(fp); err != nil {
					return err
				}
			} else {
				log.Debug().Str("path", path.Location).
					Str("desc", wm.desc).
					Uint64("rank", uint64(wm.work.Rank)).
					Msg("no match, ignoring work")
			}
		}

		return nil
	}

	return Work{WorkFunc: workFunc}, nil
}

func createMD5File(path string, md5sum []byte) (err error) { // NRV
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		err = errors.Wrap(err, "will not overwrite an existing MD5 file")
		return
	}

	defer func() {
		err = utilities.CombineErrors(err, f.Close())
	}()

	encoded := make([]byte, hex.EncodedLen(len(md5sum)))
	hex.Encode(encoded, md5sum)
	_, err = f.Write(append(encoded, '\n'))

	return
}

func translatePath(lBase string, rBase string, path FilePath) (string, error) {
	src, err := filepath.Rel(lBase, path.Location)
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(rBase, src)), err
}
