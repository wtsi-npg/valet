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
 * @file filepath_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFilePath(t *testing.T) {
	// Directory
	dirPath :="testdata/valet/testdir"
	dir, derr := NewFilePath(dirPath)
	assert.NoError(t, derr, "expected to create directory path")

	absDir, _ := filepath.Abs(dirPath)
	assert.Equal(t, dir.Location, absDir)
	assert.NotNil(t, dir.Info, "expected Info to be populated")

	// File
	fqPath := "testdata/valet/1/reads/fastq/reads1.fastq"
	file, ferr := NewFilePath(fqPath)
	assert.NoError(t, ferr, "expected to create file path")

	absFile, _ := filepath.Abs(fqPath)
	assert.Equal(t, file.Location, absFile)
	assert.NotNil(t, file.Info, "expected Info to be populated")

	// Absent
	_, nerr := NewFilePath("no_such_path")
	assert.Error(t, nerr, "expected an error for non-existent path")
}

func TestFilePath_ChecksumFilename(t *testing.T) {
	file, _ := NewFilePath("testdata/valet/1/reads/fastq/reads1.fastq")

	absDir, _ := filepath.Abs(".")
	path, err := filepath.Rel(absDir, file.ChecksumFilename())
	if assert.NoError(t, err) {
		assert.Equal(t, "testdata/valet/1/reads/fastq/reads1.fastq.md5", path)
	}
}

func TestFilePath_CompressedFilename(t *testing.T) {
	file, _ := NewFilePath("testdata/valet/1/reads/fastq/reads1.fastq")

	absDir, _ := filepath.Abs(".")
	path, err := filepath.Rel(absDir, file.CompressedFilename())
	if assert.NoError(t, err) {
		assert.Equal(t, "testdata/valet/1/reads/fastq/reads1.fastq.gz", path)
	}
}

func TestFilePath_UncompressedFilename(t *testing.T) {
	uncomp, _ := NewFilePath("testdata/valet/1/reads/fastq/reads2.fastq")
	assert.Equal(t, uncomp.UncompressedFilename(), uncomp.Location)

	comp, _ := NewFilePath("testdata/valet/1/reads/fastq/reads2.fastq.gz")
	assert.Equal(t, comp.UncompressedFilename(), uncomp.Location)
}
