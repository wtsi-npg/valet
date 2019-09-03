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
 * @file valet_suite_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kjsanger/valet/cmd"
	"github.com/kjsanger/valet/utilities"
	"github.com/kjsanger/valet/valet"
)

var dataFiles = []string{
	"./testdata/valet/1/reads/fast5/reads1.fast5",
	"./testdata/valet/1/reads/fast5/reads1.fast5.md5",
	"./testdata/valet/1/reads/fast5/reads2.fast5",
	"./testdata/valet/1/reads/fast5/reads3.fast5",
	"./testdata/valet/1/reads/fastq/reads1.fastq",
	"./testdata/valet/1/reads/fastq/reads1.fastq.md5",
	"./testdata/valet/1/reads/fastq/reads2.fastq",
	"./testdata/valet/1/reads/fastq/reads3.fastq",
	"./testdata/valet/testdir/.gitignore",
}
var dataDirs = []string{
	"./testdata/valet",
	"./testdata/valet/1",
	"./testdata/valet/1/reads",
	"./testdata/valet/1/reads/fast5",
	"./testdata/valet/1/reads/fastq",
	"./testdata/valet/testdir",
}

var _ = Describe("Find directories)", func() {
	var expectedPaths = dataDirs
	var foundDirs []valet.FilePath

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "./testdata/valet",
			valet.IsDir, valet.IsFalse)

		for p := range paths {
			foundDirs = append(foundDirs, p)
		}

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}

		cancel()
	})

	Context("when using a directory predicate", func() {
		It("should find directories", func() {
			Expect(len(foundDirs)).To(Equal(len(expectedPaths)))

			for i, ep := range expectedPaths {
				a, err := filepath.Abs(ep)
				Expect(err).NotTo(HaveOccurred())

				fp, _ := valet.NewFilePath(a)

				Expect(foundDirs[i].Location).To(Equal(fp.Location))
				Expect(foundDirs[i].Info).ToNot(BeNil())
			}
		})
	})
})

var _ = Describe("Find regular files)", func() {
	var expectedPaths = dataFiles
	var foundFiles []valet.FilePath

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "./testdata/valet",
			valet.IsRegular, valet.IsFalse)

		for p := range paths {
			foundFiles = append(foundFiles, p)
		}

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}

		cancel()
	})

	Context("when using a file predicate", func() {
		It("should find files", func() {

			Expect(len(foundFiles)).To(Equal(len(expectedPaths)))

			for i, ep := range expectedPaths {
				a, err := filepath.Abs(ep)
				Expect(err).NotTo(HaveOccurred())

				fp, _ := valet.NewFilePath(a)

				Expect(foundFiles[i].Location).To(Equal(fp.Location))
				Expect(foundFiles[i].Info).ToNot(BeNil())
			}
		})
	})
})

var _ = Describe("Find files with pruning", func() {
	var expectedPaths = []string{
		"./testdata/valet",
		"./testdata/valet/1",
		"./testdata/valet/testdir",
	}

	var foundDirs []valet.FilePath

	pruneFn := func(pf valet.FilePath) (bool, error) {
		pattern, err := filepath.Abs("./testdata/valet/1/reads")
		if err != nil {
			return false, err
		}

		match, err := filepath.Match(pattern, pf.Location)
		if err == nil && match {
			return match, filepath.SkipDir
		}

		return match, err
	}

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "./testdata/valet",
			valet.IsDir, pruneFn)

		for p := range paths {
			foundDirs = append(foundDirs, p)
		}

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}

		cancel()
	})

	Context("when using a prune function", func() {
		It("should find paths, except those pruned", func() {
			Expect(len(foundDirs)).To(Equal(len(expectedPaths)))

			for i, ep := range expectedPaths {
				a, err := filepath.Abs(ep)
				Expect(err).NotTo(HaveOccurred())

				fp, _ := valet.NewFilePath(a)

				Expect(foundDirs[i].Location).To(Equal(fp.Location))
				Expect(foundDirs[i].Info).ToNot(BeNil())
			}
		})
	})
})

var _ = Describe("Find files at intervals", func() {
	var expectedPaths = dataFiles[:len(dataFiles) -1]
	var foundFiles = map[string]bool{}

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 500 * time.Millisecond

		paths, errs := valet.FindFilesInterval(cancelCtx, "./testdata",
			valet.IsRegular, valet.IsFalse, interval)

		// Find files or timeout and cancel
		done := make(chan bool, 2)

		go func() {
			defer cancel()

			timer := time.NewTimer(5 * interval)
			<-timer.C
			done <- true // Timeout
		}()

		go func() {
			defer cancel()

			foundFiles = make(map[string]bool) // FilePaths are not comparable
			for p := range paths {
				foundFiles[p.Location] = true
				if len(foundFiles) >= len(expectedPaths) {
					done <- true // Find files
					return
				}
			}
		}()

		<-done

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}
	})

	Context("when using a file predicate", func() {
		It("should find files", func() {
			Expect(len(foundFiles)).Should(Equal(len(expectedPaths)))

			for _, ep := range expectedPaths {
				a, err := filepath.Abs(ep)
				Expect(err).NotTo(HaveOccurred())
				Expect(foundFiles[a]).To(BeTrue())
			}
		})
	})
})

var _ = Describe("Watch for file changes", func() {
	var expectedPaths = []string{
		"./testdata/valet/1/reads/fast5/reads1.fast5",
		"./testdata/valet/1/reads/fast5/reads2.fast5",
		"./testdata/valet/1/reads/fast5/reads3.fast5",
		"./testdata/valet/1/reads/fastq/reads1.fastq",
		"./testdata/valet/1/reads/fastq/reads2.fastq",
		"./testdata/valet/1/reads/fastq/reads3.fastq",
	}
	var expectedDirs = []string{
		"./testdata/valet/1/reads/fast5/",
		"./testdata/valet/1/reads/fastq/",
	}

	var foundFiles = map[string]bool{}

	var tmpDir string

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 500 * time.Millisecond

		td, terr := ioutil.TempDir("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td

		// Set up dirs to watch first
		derr := mkdirAllRelative(tmpDir, expectedDirs)
		Expect(derr).NotTo(HaveOccurred())

		paths, errs :=
			valet.WatchFiles(cancelCtx, tmpDir, valet.IsRegular, valet.IsFalse)

		cerr := copyFilesRelative(tmpDir, expectedPaths, moveFile)
		Expect(cerr).NotTo(HaveOccurred())

		// Detect updated files or timeout and cancel
		done := make(chan bool, 2)

		go func() {
			defer cancel()

			timer := time.NewTimer(5 * interval)
			<-timer.C
			done <- true // Timeout
		}()

		go func() {
			defer cancel()

			foundFiles = make(map[string]bool) // FilePaths are not comparable
			for p := range paths {
				foundFiles[p.Location] = true
				if len(foundFiles) >= len(expectedPaths) {
					done <- true // Detect files
					return
				}
			}
		}()

		<-done

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when using a file predicate", func() {
		It("should find files", func() {
			Expect(len(foundFiles)).Should(Equal(len(expectedPaths)))

			for _, ep := range expectedPaths {
				a := filepath.Join(tmpDir, ep)
				Expect(foundFiles[a]).To(BeTrue())
			}
		})
	})
})

var _ = Describe("Watch for file changes with pruning", func() {
	var allPaths = []string{
		"./testdata/valet/1/reads/fast5/reads1.fast5",
		"./testdata/valet/1/reads/fast5/reads2.fast5",
		"./testdata/valet/1/reads/fast5/reads3.fast5",
		"./testdata/valet/1/reads/fastq/reads1.fastq",
		"./testdata/valet/1/reads/fastq/reads2.fastq",
		"./testdata/valet/1/reads/fastq/reads3.fastq",
	}
	var allDirs = []string{
		"./testdata/valet/1/reads/fast5/",
		"./testdata/valet/1/reads/fastq/",
	}

	var tmpDir string

	expectedPaths := allPaths[:4]

	var foundFiles = map[string]bool{}

	pruneFn := func(pf valet.FilePath) (bool, error) {
		pattern, err := filepath.Abs("./testdata/valet/1/reads/fastq")
		if err != nil {
			return false, err
		}

		match, err := filepath.Match(pattern, pf.Location)
		if err == nil && match {
			fmt.Println(fmt.Sprintf("matched %s", pf.Location))
			return match, filepath.SkipDir
		}

		return match, err
	}

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 1 * time.Second

		td, terr := ioutil.TempDir("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td

		// Set up dirs to watch first
		derr := mkdirAllRelative(tmpDir, allDirs)
		Expect(derr).NotTo(HaveOccurred())

		paths, errs :=
			valet.WatchFiles(cancelCtx, tmpDir, valet.IsRegular, pruneFn)

		cerr := copyFilesRelative(tmpDir, allPaths, moveFile)
		Expect(cerr).NotTo(HaveOccurred())

		// Detect updated files or timeout and cancel
		done := make(chan bool, 2)

		go func() {
			defer cancel()

			timer := time.NewTimer(5 * interval)
			<-timer.C
			done <- true // Timeout
		}()

		go func() {
			defer cancel()

			foundFiles = make(map[string]bool)
			for p := range paths {
				foundFiles[p.Location] = true
				if len(foundFiles) >= len(expectedPaths) {
					done <- true // Detect files
					return
				}
			}
		}()

		<-done

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when using a file predicate", func() {
		It("should find files", func() {
			Expect(len(foundFiles)).Should(Equal(len(expectedPaths)))

			for _, ep := range expectedPaths {
				a := filepath.Join(tmpDir, ep)
				Expect(foundFiles[a]).To(BeTrue())
			}
		})
	})
})

var _ = Describe("Count files without a checksum", func() {
	var numFilesFound uint64
	var numFilesExpected uint64 = 4

	BeforeEach(func() {
		n, err := cmd.CountFilesWithoutChecksum("./testdata/valet", []string{})
		Expect(err).NotTo(HaveOccurred())
		numFilesFound = n
	})

	Context("when there are data files without checksum files", func() {
		It("should count those files", func() {
			Expect(numFilesFound).Should(Equal(numFilesExpected))
		})
	})
})

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

func copyFilesRelative(root string, relPaths []string, fn copyFn) error {
	for _, p := range relPaths {
		from, err := filepath.Abs(p)
		if err != nil {
			return err
		}

		to := filepath.Join(root, p)

		err = fn(from, to)
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
