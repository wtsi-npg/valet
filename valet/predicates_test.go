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
 * @file predicates_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kjsanger/valet/utilities"
	"github.com/stretchr/testify/assert"

	logs "github.com/kjsanger/logshim"
	"github.com/kjsanger/logshim/dlog"
)

func init() {
	log := dlog.New(os.Stderr, logs.ErrorLevel)
	logs.InstallLogger(log)
}

func TestIsDir(t *testing.T) {
	fp, _ := NewFilePath("./testdata/testdir")

	ok, err := IsDir(fp)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a directory")
	}
}

func TestIsRegular(t *testing.T) {
	fq, _ := NewFilePath("./testdata/1/reads/fastq/reads1.fastq")
	ok, err := IsRegular(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a file")
	}
}

func TestIsFast5Match(t *testing.T) {
	f5, _ := NewFilePath("./testdata/1/reads/fast5/reads1.fast5")
	ok, err := IsFast5Match(f5)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file")
	}
}

func TestIsFastqMatch(t *testing.T) {
	fq, _ := NewFilePath("./testdata/1/reads/fastq/reads1.fastq")
	ok, err := IsFastqMatch(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fastq file")
	}
}

func TestHasChecksumFile(t *testing.T) {
	f5With, _ := NewFilePath("./testdata/1/reads/fast5/reads1.fast5")
	ok, err := HasChecksumFile(f5With)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file with checksum")
	}

	f5Without, _ := NewFilePath("./testdata/1/reads/fast5/reads2.fast5")
	ok, err = HasChecksumFile(f5Without)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a fast5 file without checksum")
	}

	fqWith, _ := NewFilePath("./testdata/1/reads/fastq/reads1.fastq")
	ok, err = HasChecksumFile(fqWith)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fastq file with checksum")
	}

	fqWithout, _ := NewFilePath("./testdata/1/reads/fastq/reads2.fastq")
	ok, err = HasChecksumFile(fqWithout)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a fastq file without checksum")
	}
}

func TestHasStaleChecksumFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestHasStaleChecksumFile")
	defer os.RemoveAll(tmpDir)
	assert.NoError(t, err)

	dataFile, checkSumFile :=
		filepath.Join(tmpDir, "reads1.fast5"),
		filepath.Join(tmpDir, "reads1.fast5.md5")

	// First write the checksum file
	err = utilities.CopyFile("./testdata/1/reads/fast5/reads1.fast5.md5",
		checkSumFile, 0600)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Then update with a newer reads file
	err = utilities.CopyFile("./testdata/1/reads/fast5/reads1.fast5",
		dataFile, 0600)
	assert.NoError(t, err)

	f5With, _ := NewFilePath(dataFile)
	ok, err := HasChecksumFile(f5With)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file with checksum")
	}

	ok, err = HasStaleChecksumFile(f5With)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file with stale checksum")
	}
}
