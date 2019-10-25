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
 * @file workfunc_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	logs "github.com/kjsanger/logshim"
	"github.com/kjsanger/logshim-zerolog/zlog"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/kjsanger/valet/utilities"
)

func TestMain(m *testing.M) {
	loggerImpl := zlog.New(os.Stderr, logs.ErrorLevel)

	writer := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	consoleLogger := loggerImpl.Logger.Output(zerolog.SyncWriter(writer))
	loggerImpl.Logger = &consoleLogger
	logs.InstallLogger(loggerImpl)

	os.Exit(m.Run())
}

func TestDoNothing(t *testing.T) {
	path, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	assert.NoError(t, DoNothing(path))
}

func TestCreateMD5ChecksumFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestCreateMD5ChecksumFile")
	defer os.RemoveAll(tmpDir)
	assert.NoError(t, err)

	dataFile, checkSumFile :=
		filepath.Join(tmpDir, "reads1.fast5"),
		filepath.Join(tmpDir, "reads1.fast5.md5")

	// First write the data file
	err = utilities.CopyFile("./testdata/valet/1/reads/fast5/reads1.fast5",
		dataFile, 0600)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	path, _ := NewFilePath(dataFile)
	err = CreateMD5ChecksumFile(path)

	if assert.NoError(t, err) {
		assert.FileExists(t, checkSumFile)
	}
}

func TestReadMD5ChecksumFile(t *testing.T) {
	f, err := NewFilePath("testdata/valet/1/reads/fast5/reads1.fast5.md5")
	assert.NoError(t, err)

	md5sum, err := ReadMD5ChecksumFile(f)
	assert.NoError(t, err)
	assert.Equal(t, "1181c1834012245d785120e3505ed169", string(md5sum))
}

func TestRemoveMD5ChecksumFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestCreateMD5ChecksumFile")
	defer os.RemoveAll(tmpDir)
	assert.NoError(t, err)

	dataFile, checkSumFile :=
		filepath.Join(tmpDir, "reads1.fast5"),
		filepath.Join(tmpDir, "reads1.fast5.md5")

	// First write the file
	err = utilities.CopyFile("./testdata/valet/1/reads/fast5/reads1.fast5",
		dataFile, 0600)
	assert.NoError(t, err)
	err = utilities.CopyFile("./testdata/valet/1/reads/fast5/reads1.fast5.md5",
		checkSumFile, 0600)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	path, _ := NewFilePath(dataFile)
	err = RemoveMD5ChecksumFile(path)

	if assert.NoError(t, err) {
		assert.FileExists(t, dataFile)

		_, err := os.Lstat(checkSumFile)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	}
}

func TestCalculateFileMD5(t *testing.T) {
	path, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")

	md5sum, err := CalculateFileMD5(path)

	if assert.NoError(t, err) {
		encoded := make([]byte, hex.EncodedLen(len(md5sum)))
		hex.Encode(encoded, md5sum)
		assert.Equal(t, string(encoded), "5c9597f3c8245907ea71a89d9d39d08e")
	}
}

func TestCompressFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestCompressFile")
	defer os.RemoveAll(tmpDir)
	assert.NoError(t, err)

	tmpFile := filepath.Join(tmpDir, "TestCompressFile")
	of := fmt.Sprintf("of=%s", tmpFile)

	// Makes a new random file each time
	err = exec.Command("dd", "if=/dev/urandom", of,
		"bs=1M", "count=50").Run()
	assert.NoError(t, err)

	randFile, err := NewFilePath(tmpFile)
	assert.NoError(t, err)

	// Calculates the MD5 of the uncompressed file
	err = CreateMD5ChecksumFile(randFile)
	assert.NoError(t, err)

	checksumFile, err := NewFilePath(randFile.ChecksumFilename())
	assert.NoError(t, err)

	expectedMD5, err := ReadMD5ChecksumFile(checksumFile)
	assert.NoError(t, err)

	// Compress and check result
	err = CompressFile(randFile)
	assert.NoError(t, err)

	compFile, err := NewFilePath(randFile.CompressedFilename())
	assert.NoError(t, err)

	err = compressedFileMatches(compFile.Location, expectedMD5)
	assert.NoError(t, err)
}

func compressedFileMatches(path string, md5sum []byte) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return err
	}

	reader, err := gzip.NewReader(f)
	defer reader.Close()
	if err != nil {
		return err
	}

	h := md5.New()
	if _, err = io.Copy(h, reader); err != nil {
		return err
	}

	sum := h.Sum(nil)
	encoded := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(encoded, sum)

	if string(encoded) != string(md5sum) {
		return errors.Errorf("MD5 %s does not match expected MD5 %s",
			encoded, md5sum)
	}

	return err
}
