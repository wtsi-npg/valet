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
 * @file pathproc.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	logf "valet/log/logfacade"
	"valet/utilities"
)

type token struct{}
type semaphore chan token

// WorkFunc is a worker function used by ProcessFiles.
type WorkFunc func(path FilePath) error

// ProcessFiles operates by applying workFunc to each FilePath in the paths
// channel. An error is raised if any workFunc itself encounters an error.
func ProcessFiles(paths <-chan FilePath, workFunc WorkFunc, maxThreads int) error {
	var wg sync.WaitGroup
	var jobCounter uint64
	var errCounter uint64
	sem := make(semaphore, maxThreads)

	log := logf.GetLogger()

	for path := range paths {
		wg.Add(1)
		sem <- token{}

		go func(p FilePath) {
			defer func() {
				<-sem
				wg.Done()
			}()

			atomic.AddUint64(&jobCounter, 1)
			log.Info().Str("path", p.Location).Msg("starting work")

			err := workFunc(p)
			if err != nil {
				log.Error().Err(err).
					Str("path", p.Location).
					Msg("worker function failed")
				atomic.AddUint64(&errCounter, 1)
			}
		}(path)
	}

	jobCount := atomic.LoadUint64(&jobCounter)
	errCount := atomic.LoadUint64(&errCounter)
	if errCount > 0 {
		return errors.Errorf("encountered %d errors processing %d files",
			errCount, jobCount)
	}

	log.Info().Uint64("num_files", jobCount).Msg("finished processing")

	return nil
}

func DoNothing(path FilePath) error {
	if path.Info.IsDir() {
		return nil
	}
	log := logf.GetLogger()
	log.Info().Str("path", path.Location).Msg("work started")
	time.Sleep(5 * time.Second)
	log.Info().Str("path", path.Location).Msg("work done")
	return nil
}

// RecordChecksum calculates a checksum for the file at path and writes it to a
// new checksum file as a hex-encoded string. It raises an error if the
// checksum file already exists.
func RecordChecksum(path FilePath) error {
	md5sum, err := CalculateFileMD5(path)
	if err != nil {
		return err
	}

	return createMD5File(ChecksumFilename(path), md5sum)
}

// CalculateFileMD5 returns the MD5 checksum of the file at path.
func CalculateFileMD5(path FilePath) ([16]byte, error) {
	var md5sum [16]byte

	f, err := os.Open(path.Location)
	defer func() {
		err = utilities.CombineErrors(err, f.Close())
	}()

	if err != nil {
		return md5sum, err
	}

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return md5sum, err
	}

	return md5.Sum(nil), err
}

func createMD5File(path string, md5sum [16]byte) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	defer func() {
		err = utilities.CombineErrors(err, f.Close())
	}()

	if err != nil {
		return errors.Wrap(err, "will not overwrite an existing MD5 file")
	}

	encoded := make([]byte, hex.EncodedLen(len(md5sum)))
	hex.Encode(encoded, md5sum[:16])
	_, err = f.Write(append(encoded, '\n'))

	return err
}
