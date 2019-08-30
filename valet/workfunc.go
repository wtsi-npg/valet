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
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"sort"

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

// WorkArr is a series of Work to be executed in order.
type WorkArr []Work

func (s WorkArr) Len() int {
	return len(s)
}

func (s WorkArr) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s WorkArr) Less(i, j int) bool {
	return s[i].Rank > s[j].Rank
}

func (s WorkArr) IsEmpty() bool {
	return len(s) == 0
}

// WorkMatch is an association between a FilePredicate and Work to be done. If
// the predicate returns true then the work will be done.
type WorkMatch struct {
	pred FilePredicate // Predicate to match against candidate FilePath
	work Work          // Work to be executed on a matching FilePath.
}

// WorkPlan is a slice of WorkMatches. Where more than one Work is matched,
// they will be done in rank order.
type WorkPlan []WorkMatch

// DryRunWorkPlan matches any FilePath and does DoNothing Work.
func DryRunWorkPlan() WorkPlan {
	var plan []WorkMatch
	plan = append(plan, WorkMatch{IsTrue, Work{WorkFunc: DoNothing}})
	return plan
}

// CreateChecksumWorkPlan manages checksum files.
func CreateChecksumWorkPlan() WorkPlan {
	var plan []WorkMatch
	plan = append(plan, WorkMatch{RequiresChecksum,
		Work{WorkFunc: CreateOrUpdateMD5ChecksumFile}})
	return plan
}

// ChecksumStateWorkPlan counts files that do not have checksums.
func ChecksumStateWorkPlan(countFunc WorkFunc) WorkPlan {
	var plan []WorkMatch
	plan = append(plan, WorkMatch{RequiresChecksum,
		Work{WorkFunc: countFunc}})
	return plan
}

// DispatchWork accepts a candidate FilePath and a WorkPlan and returns Work
// encapsulating all the work in the WorkPlan. If no work is required for the
// FilePath, it returns DoNothing Work.
func DispatchWork(path FilePath, plan WorkPlan) (Work, error) {
	var work Work

	var matchedWork WorkArr
	for _, m := range plan {
		add, err := m.pred(path)
		if err != nil {
			return work, err
		}
		if add {
			matchedWork = append(matchedWork, m.work)
		}
	}
	work = combineWork(matchedWork)

	return work, nil
}

func combineWork(work WorkArr) Work {
	var workFunc WorkFunc

	if work.IsEmpty() {
		workFunc = DoNothing
	} else {
		sort.Sort(work)

		workFunc = func(path FilePath) error {
			for _, w := range work {
				if err := w.WorkFunc(path); err != nil {
					return err
				}
			}
			return nil
		}
	}

	return Work{WorkFunc: workFunc}
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

	ok, err := HasStaleChecksumFile(path)
	if err != nil {
		log.Error().Err(err).Msg("stale checksum detection failed")
		return err
	}

	if ok {
		return UpdateMD5ChecksumFile(path)
	}

	ok, err = HasChecksumFile(path)
	if err != nil {
		return err
	}

	if !ok {
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
func CalculateFileMD5(path FilePath) (md5sum []byte, err error) {
	f, err := os.Open(path.Location)
	if err != nil {
		return md5sum, err
	}

	defer func() {
		err = utilities.CombineErrors(err, f.Close())
	}()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return md5sum, err
	}

	return h.Sum(nil), err
}

func createMD5File(path string, md5sum []byte) (err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return errors.Wrap(err, "will not overwrite an existing MD5 file")
	}

	defer func() {
		err = utilities.CombineErrors(err, f.Close())
	}()

	encoded := make([]byte, hex.EncodedLen(len(md5sum)))
	hex.Encode(encoded, md5sum)
	_, err = f.Write(append(encoded, '\n'))

	return err
}
