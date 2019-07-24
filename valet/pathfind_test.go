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
 * @file pathfind_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"os"
	"path/filepath"
	"testing"

	logs "github.com/kjsanger/logshim"
	"github.com/kjsanger/logshim/dlog"

	"github.com/stretchr/testify/assert"
)

func init() {
	log := dlog.New(os.Stderr, logs.ErrorLevel)
	logs.InstallLogger(log)
}

func TestNewFilePath(t *testing.T) {
	dir, derr := NewFilePath("./testdata/testdir")
	assert.NoError(t, derr, "expected to create directory path")

	absDir, _ := filepath.Abs("./testdata/testdir")
	assert.Equal(t, dir.Location, absDir)
	assert.NotNil(t, dir.Info, "expected Info to be populated")

	file, ferr := NewFilePath("./testdata/1/reads/fastq/reads1.fastq")
	assert.NoError(t, ferr, "expected to create file path")

	absFile, _ := filepath.Abs("./testdata/1/reads/fastq/reads1.fastq")
	assert.Equal(t, file.Location, absFile)
	assert.NotNil(t, file.Info, "expected Info to be populated")

	_, nerr := NewFilePath("./no such path")
	assert.Error(t, nerr, "expected an error for non-existent path")
}
