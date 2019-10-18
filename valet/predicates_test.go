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

	"github.com/stretchr/testify/assert"

	"github.com/kjsanger/valet/utilities"
)

func TestIsDir(t *testing.T) {
	fp, _ := NewFilePath("./testdata/valet/testdir")

	ok, err := IsDir(fp)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a directory")
	}
}

func TestIsRegular(t *testing.T) {
	fq, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err := IsRegular(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a file")
	}
}

// mockFileStat is an implementation of FileInfo that lets you set a size.
type mockFileStat struct {
	size int64
}

func (fs *mockFileStat) Name() string       { return "" }
func (fs *mockFileStat) Size() int64        { return fs.size }
func (fs *mockFileStat) Mode() os.FileMode  { return 0 }
func (fs *mockFileStat) ModTime() time.Time { return time.Now() }
func (fs *mockFileStat) IsDir() bool        { return false }
func (fs *mockFileStat) Sys() interface{}   { return nil }

func TestIsLarge(t *testing.T) {
	fq, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err := IsLarge(fq)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a file")
	}

	fq.Info = &mockFileStat{size: 524288000}
	ok, err = IsLarge(fq)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a file")
	}

	fq.Info = &mockFileStat{size: 524288001}
	ok, err = IsLarge(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a file")
	}
}

func TestIsFast5Match(t *testing.T) {
	f5, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads1.fast5")
	ok, err := IsFast5(f5)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file")
	}
}

func TestIsFastqMatch(t *testing.T) {
	fq, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err := IsFastq(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fastq file")
	}
}

func TestIsCompressedMatch(t *testing.T) {
	fq, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq.gz")
	ok, err := IsCompressed(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a gz file")
	}

	fq, _ = NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err = IsCompressed(fq)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a fastq file")
	}
}

func TestHasChecksumFile(t *testing.T) {
	f5With, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads1.fast5")
	ok, err := HasChecksumFile(f5With)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file with checksum")
	}

	f5Without, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads2.fast5")
	ok, err = HasChecksumFile(f5Without)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a fast5 file without checksum")
	}

	f5WithCompressed, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads3.fast5")
	ok, err = HasChecksumFile(f5WithCompressed)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file with checksum on the compressed file")
	}

	fqWith, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err = HasChecksumFile(fqWith)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fastq file with checksum")
	}

	fqWithout, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads2.fastq")
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
	err = utilities.CopyFile("./testdata/valet/1/reads/fast5/reads1.fast5.md5",
		checkSumFile, 0600)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Then update with a newer reads file
	err = utilities.CopyFile("./testdata/valet/1/reads/fast5/reads1.fast5",
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

func TestHasCompressedVersion(t *testing.T) {
	f5With, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads3.fast5")
	ok, err := HasCompressedVersion(f5With)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file with gz")
	}

	f5Without, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads1.fast5")
	ok, err = HasChecksumFile(f5Without)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a fast5 file without gz")
	}
}

func TestHasStaleCompressionFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestHasStaleCompressionFile")
	defer os.RemoveAll(tmpDir)
	assert.NoError(t, err)

	dataFile, compressedFile :=
		filepath.Join(tmpDir, "reads1.fast5"),
		filepath.Join(tmpDir, "reads1.fast5.gzip")

	// First write the compressed file
	err = utilities.CopyFile("./testdata/valet/1/reads/fast5/reads1.fast5.gzip",
		compressedFile, 0600)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Then update with a newer reads file
	err = utilities.CopyFile("./testdata/valet/1/reads/fast5/reads1.fast5",
		dataFile, 0600)
	assert.NoError(t, err)

	f5With, _ := NewFilePath(dataFile)
	ok, err := HasCompressedVersion(f5With)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file with gzip")
	}

	ok, err = HasStaleCompressedFile(f5With)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file with stale compressed file")
	}
}

func TestIsMinKNOWRunDir(t *testing.T) {
	gridionRunDirs := []string{
		"testdata/platform/ont/minknow/gridion/66/DN585561I_A1/" +
			"20190904_1514_GA20000_FAL01979_43578c8f",
		"testdata/platform/ont/minknow/gridion/66/DN585561I_B1/" +
			"20190904_1514_GA30000_FAL09731_2f0f08bc",
	}

	promethionDirs := []string{
		"testdata/platform/ont/minknow/promethion/DN467851H_Multiplex_Pool_1/" +
			"DN467851H_B2_C2_E2_F2/20190820_1538_2-E7-H7_PAD71219_a4a384ec",
		"testdata/platform/ont/minknow/promethion/DN467851H_Multiplex_Pool_2/" +
			"DN467851H_A3_F3_G3_H3/20190821_1545_1-A1-D1_PAD73195_440ab859",
	}

	var allRunDirs []string
	allRunDirs = append(allRunDirs, gridionRunDirs...)
	allRunDirs = append(allRunDirs, promethionDirs...)

	for _, dir := range allRunDirs {
		fp, _ := NewFilePath(dir)
		ok, err := IsMinKNOWRunDir(fp)
		if assert.NoError(t, err) {
			assert.True(t, ok, "expected %s to be a GridION run directory", dir)
		}
	}
}
