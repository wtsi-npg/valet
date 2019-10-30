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
 * @file filepath.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileResource is a locatable file.
type FileResource struct {
	Location string // Raw URL or file path
}

// FilePath is a FileResource that is on a local filesystem. It represents a
// file that is present at the time the instance is created.
type FilePath struct {
	FileResource
	Info os.FileInfo
}

// NewFilePath returns a new instance where the path has been cleaned and made
// absolute and the FileInfo populated by os.Stat. FilePaths should be
// created using this constructor to ensure that always have a clean, absolute
// path and populated FileInfo.
func NewFilePath(path string) (FilePath, error) {
	var fp FilePath
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return fp, err
	}

	info, err := os.Stat(absPath)
	fp.Info = info
	fp.FileResource = FileResource{absPath}

	return fp, err
}

// ChecksumFilename returns the expected path of the checksum file belonging
// to the path.
func (path *FilePath) ChecksumFilename() string {
	return fmt.Sprintf("%s.%s", path.Location, MD5Suffix)
}

// CompressedFilename returns the expected path of the compressed  version of
// this file.
func (path *FilePath) CompressedFilename() string {
	return fmt.Sprintf("%s.%s", path.Location, GzipSuffix)
}

// UncompressedFilename returns the expected path of the uncompressed version
// of this file. If the file is not compressed, returns the path of this file.
func (path *FilePath) UncompressedFilename() string {
	const dotGz = "." + GzipSuffix
	return strings.TrimSuffix(path.Location, dotGz)
}
