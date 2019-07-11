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
 * @file pathproc_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	logf "valet/log/logfacade"
	logs "valet/log/slog"
)

func init() {
	log := logs.New(os.Stderr, logf.ErrorLevel)
	logf.InstallLogger(log)
}

func TestCalculateFileMD5(t *testing.T) {
	file, ferr := NewFilePath("./testdata/1/reads/fastq/reads1.fastq")
	assert.NoError(t, ferr, "expected to create file path")

	md5sum, err := CalculateFileMD5(file)
	encoded := make([]byte, hex.EncodedLen(len(md5sum)))
	hex.Encode(encoded, md5sum)

	if assert.NoError(t, err) {
		assert.Equal(t, string(encoded), "5c9597f3c8245907ea71a89d9d39d08e")
	}
}
