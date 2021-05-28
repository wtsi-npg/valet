/*
 * Copyright (C) 2019, 2021. Genome Research Ltd. All rights reserved.
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

	"github.com/wtsi-npg/valet/utilities"
)

func TestIsDir(t *testing.T) {
	d, _ := NewFilePath("./testdata/valet/testdir")
	ok, err := IsDir(d)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a directory")
	}

	f, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err = IsDir(f)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a file")
	}
}

func TestIsRegular(t *testing.T) {
	f, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err := IsRegular(f)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a file")
	}

	d, _ := NewFilePath("./testdata/valet/testdir")
	ok, err = IsRegular(d)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a directory")
	}
}

func TestIsFast5Match(t *testing.T) {
	f5, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads1.fast5")
	ok, err := IsFast5(f5)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fast5 file")
	}

	fq, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err = IsFast5(fq)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a non-fast5 file")
	}
}

func TestIsFastqMatch(t *testing.T) {
	fq, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok, err := IsFastq(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a fastq file")
	}

	f5, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads1.fast5")
	ok, err = IsFastq(f5)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a non-fastq file")
	}
}

func TestIsBAMMatch(t *testing.T) {
	bam, _ := NewFilePath("./testdata/valet/1/reads/alignments/alignments1.bam")
	ok, err := IsBAM(bam)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a BAM file")
	}

	f5, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads1.fast5")
	ok, err = IsBAM(f5)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a non-BAM file")
	}
}

func TestIsBEDMatch(t *testing.T) {
	bed, _ := NewFilePath("./testdata/valet/1/adaptive_sampling_roi1.bed")
	ok, err := IsBED(bed)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a BED file")
	}

	f5, _ := NewFilePath("./testdata/valet/1/reads/fast5/reads1.fast5")
	ok, err = IsBED(f5)
	if assert.NoError(t, err) {
		assert.False(t, ok, "expected false for a non-BED file")
	}
}

func TestIsGzipFastqMatch(t *testing.T) {
	fq, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads2.fastq.gz")

	ok, err := IsCompressed(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a gzipped fastq file")
	}

	ok, err = IsFastq(fq)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a gzipped fastq file")
	}
}

func TestIsGzipCSVMatch(t *testing.T) {
	csv, _ := NewFilePath("./testdata/valet/1/ancillarey.csv.gz")

	ok, err := IsCompressed(csv)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a gzipped CSV file")
	}

	ok, err = IsCSV(csv)
	if assert.NoError(t, err) {
		assert.True(t, ok, "expected true for a gzipped CSV file")
	}
}

func TestIsCompressed(t *testing.T) {
	fq1, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	ok1, err1 := IsCompressed(fq1)
	if assert.NoError(t, err1) {
		assert.False(t, ok1, "expected false for a fastq file")
	}

	fq2, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads2.fastq.gz")

	ok2, err2 := IsCompressed(fq2)
	if assert.NoError(t, err2) {
		assert.True(t, ok2, "expected true for a gzipped fastq file")
	}
}

func TestHasCompressedVersion(t *testing.T) {
	fq1, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")

	ok1, err1 := HasCompressedVersion(fq1)
	if assert.NoError(t, err1) {
		assert.False(t, ok1,
			"expected false for a fastq file without a gzipped version")
	}

	fq2, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads2.fastq")

	ok2, err2 := HasCompressedVersion(fq2)
	if assert.NoError(t, err2) {
		assert.True(t, ok2,
			"expected true for a fastq file with a gzipped version")
	}
}

func TestRequiresCompression(t *testing.T) {
	fq1, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")

	ok1, err1 := RequiresCompression(fq1)
	if assert.NoError(t, err1) {
		assert.True(t, ok1,
			"expected true for a fastq file without a gzipped version")
	}

	fq2, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads2.fastq")

	ok2, err2 := RequiresCompression(fq2)
	if assert.NoError(t, err2) {
		assert.False(t, ok2,
			"expected false for a fastq file with a gzipped version")
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
