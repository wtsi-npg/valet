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
	"strings"
	"testing"

"github.com/stretchr/testify/assert"
)

func TestNewFilePath(t *testing.T) {
	dir, derr := NewFilePath("./testdata/valet/testdir")
	assert.NoError(t, derr, "expected to create directory path")

	absDir, _ := filepath.Abs("./testdata/valet/testdir")
	assert.Equal(t, dir.Location, absDir)
	assert.NotNil(t, dir.Info, "expected Info to be populated")

	file, ferr := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	assert.NoError(t, ferr, "expected to create file path")

	absFile, _ := filepath.Abs("./testdata/valet/1/reads/fastq/reads1.fastq")
	assert.Equal(t, file.Location, absFile)
	assert.NotNil(t, file.Info, "expected Info to be populated")

	_, nerr := NewFilePath("./no such path")
	assert.Error(t, nerr, "expected an error for non-existent path")
}

func TestFilePath_ChecksumFilename(t *testing.T) {
	file, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	assert.Equal(t, file.ChecksumFilename(), file.Location + "." + MD5Suffix)
}

func TestFilePath_CompressedFilename(t *testing.T) {
	file, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads1.fastq")
	assert.Equal(t, file.CompressedFilename(), file.Location + "." + GzipSuffix)
}

func TestFilePath_UncompressedFilename(t *testing.T) {
	uncomp, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads2.fastq")
	assert.Equal(t, uncomp.UncompressedFilename(), uncomp.Location)

	comp, _ := NewFilePath("./testdata/valet/1/reads/fastq/reads2.fastq.gz")
	assert.Equal(t, comp.UncompressedFilename(),
		strings.TrimSuffix(comp.Location, "." + GzipSuffix))
}
