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
 * @file helpers_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	ex "github.com/kjsanger/extendo"
	. "github.com/onsi/ginkgo"

	"github.com/kjsanger/valet/utilities"
	"github.com/kjsanger/valet/valet"
)

type itemPathTransform func(i []ex.RodsItem) []string
type localPathTransform func(p []valet.FilePath) []string

func makeRodsItemTransform(workColl string) func(i []ex.RodsItem) []string {
	return func(items []ex.RodsItem) []string {
		var paths []string
		for _, p := range items {
			r, _ := filepath.Rel(workColl, p.RodsPath())
			paths = append(paths, r)
		}
		return paths
	}
}

func makeLocalPathTransform(root string) func(i []valet.FilePath) []string {
	return func(items []valet.FilePath) []string {
		var paths []string
		for _, p := range items {
			r, err := filepath.Rel(root, p.Location)
			if err != nil {
				panic(err)
			}
			paths = append(paths, r)
		}
		return paths
	}
}

func findFilesRelative(root string) ([]string, error) {
	return findRelative(root, func(i os.FileInfo) bool {
		return i.IsDir()
	})
}

func findDirsRelative(root string) ([]string, error) {
	return findRelative(root, func(i os.FileInfo) bool {
		return !i.IsDir()
	})
}

func findRelative(root string, filter func(i os.FileInfo) bool) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !filter(info) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// Recursively create these subdirs under root
func mkdirAllRelative(root string, subdirs []string) error {
	for _, dir := range subdirs {
		err := os.MkdirAll(filepath.Join(root, dir), 0700)
		if err != nil {
			return err
		}
	}

	return nil
}

type copyFn func(to string, from string) error

// Copy files to subdirs under root
func copyFilesRelative(from string, to string, relPaths []string, fn copyFn) error {
	fr, err := filepath.Abs(from)
	if err != nil {
		return err
	}

	for _, p := range relPaths {
		src := filepath.Join(fr, p)
		dst := filepath.Join(to, p)

		err = fn(src, dst)
		if err != nil {
			return err
		}
	}
	return nil
}

// A copyFn using Open/Write/Close
func readWriteFile(from string, to string) error {
	return utilities.CopyFile(from, to, 0600)
}

// A copyFn using os.Rename
func moveFile(from string, to string) error {
	stagingDir, err := ioutil.TempDir("", "ValetTests")
	defer os.RemoveAll(stagingDir)
	if err != nil {
		return err
	}

	stagingFile := filepath.Join(stagingDir, filepath.Base(from))
	err = readWriteFile(from, stagingFile)
	if err != nil {
		return err
	}

	return os.Rename(stagingFile, to)
}

// Remove test data recursively from under path dst from iRODS
func removeTmpCollection(dst string) error {
	client, err := ex.FindAndStart("--unbuffered")
	if err != nil {
		return err
	}
	_, err = client.RemDir(ex.Args{Force: true, Recurse: true},
		ex.RodsItem{IPath: dst})
	if err != nil {
		return err
	}

	return client.Stop()
}

// Return a new pseudo randomised path in iRODS
func tmpRodsPath(root string, prefix string) string {
	s := rand.NewSource(GinkgoRandomSeed())
	r := rand.New(s)
	d := fmt.Sprintf("%s.%d.%010d", prefix, os.Getpid(), r.Uint32())
	return filepath.Join(root, d)
}

func toArray(m map[string]valet.FilePath) ([]valet.FilePath, error) {
	var fp valet.FilePathArr
	for _, path := range m {
		fp = append(fp, path)
	}
	sort.Sort(fp)

	return fp, nil
}
