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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"valet/utilities"
	"valet/valet"

	logf "valet/log/logfacade"
	logs "valet/log/slog"
)

func TestValet(t *testing.T) {
	log := logs.New(GinkgoWriter, logf.ErrorLevel)
	logf.InstallLogger(log)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Valet Suite")
}

var _ = Describe("FindFiles/IsDir)", func() {
	var foundDirs []valet.FilePath
	var paths = []string{
		"./testdata",
		"./testdata/1",
		"./testdata/1/reads",
		"./testdata/1/reads/fast5",
		"./testdata/1/reads/fastq",
		"./testdata/testdir",
	}

	BeforeEach(func() {
		cancelCtx, _ := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "./testdata",
			valet.IsDir, valet.IsIdentity)

		for p := range paths {
			foundDirs = append(foundDirs, p)
		}

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}
	})

	Context("when using a directory predicate", func() {
		It("should find directories", func() {
			Expect(len(foundDirs)).To(Equal(len(paths)))

			for i, ep := range paths {
				a, err := filepath.Abs(ep)
				Expect(err).NotTo(HaveOccurred())

				fp, _ := valet.NewFilePath(a)

				Expect(foundDirs[i].Location).To(Equal(fp.Location))
				Expect(foundDirs[i].Info).ToNot(BeNil())
			}
		})
	})
})

var _ = Describe("FindFiles/IsRegular)", func() {
	var foundFiles []valet.FilePath
	var paths = []string{
		"./testdata/1/reads/fast5/reads1.fast5",
		"./testdata/1/reads/fast5/reads1.fast5.md5",
		"./testdata/1/reads/fast5/reads2.fast5",
		"./testdata/1/reads/fast5/reads3.fast5",
		"./testdata/1/reads/fastq/reads1.fastq",
		"./testdata/1/reads/fastq/reads1.fastq.md5",
		"./testdata/1/reads/fastq/reads2.fastq",
		"./testdata/1/reads/fastq/reads3.fastq",
		"./testdata/testdir/.gitignore",
	}

	BeforeEach(func() {
		cancelCtx, _ := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "./testdata",
			valet.IsRegular, valet.IsIdentity)

		for p := range paths {
			foundFiles = append(foundFiles, p)
		}

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}
	})

	Context("when using a file predicate", func() {
		It("should find files", func() {

			Expect(len(foundFiles)).To(Equal(len(paths)))

			for i, ep := range paths {
				a, err := filepath.Abs(ep)
				Expect(err).NotTo(HaveOccurred())

				fp, _ := valet.NewFilePath(a)

				Expect(foundFiles[i].Location).To(Equal(fp.Location))
				Expect(foundFiles[i].Info).ToNot(BeNil())
			}
		})
	})
})

var _ = Describe("FindFilesInterval", func() {
	var foundFiles []valet.FilePath
	var expectedPaths = []string{
		"./testdata/1/reads/fast5/reads1.fast5",
		"./testdata/1/reads/fast5/reads1.fast5.md5",
		"./testdata/1/reads/fast5/reads2.fast5",
		"./testdata/1/reads/fast5/reads3.fast5",
		"./testdata/1/reads/fastq/reads1.fastq",
		"./testdata/1/reads/fastq/reads1.fastq.md5",
		"./testdata/1/reads/fastq/reads2.fastq",
		"./testdata/1/reads/fastq/reads3.fastq",
	}

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 1 * time.Second

		paths, errs := valet.FindFilesInterval(cancelCtx,
			"./testdata", valet.IsRegular, interval)

		// Find files or timeout and cancel
		done := make(chan bool, 2)

		go func() {
			defer cancel()

			timer := time.NewTimer(5 * interval)
			<-timer.C
			done <- true // Timeout
			return
		}()

		go func() {
			defer cancel()

			for p := range paths {
				foundFiles = append(foundFiles, p)

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

var _ = Describe("WatchFiles", func() {
	var foundFiles []valet.FilePath
	var expectedPaths = []string{
		"./testdata/1/reads/fast5/reads1.fast5",
		"./testdata/1/reads/fast5/reads2.fast5",
		"./testdata/1/reads/fast5/reads3.fast5",
		"./testdata/1/reads/fastq/reads1.fastq",
		"./testdata/1/reads/fastq/reads2.fastq",
		"./testdata/1/reads/fastq/reads3.fastq",
	}
	var expectedDirs = []string{
		"./testdata/1/reads/fast5/",
		"./testdata/1/reads/fastq/",
	}

	var tmpDir string

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 1 * time.Second

		td, terr := ioutil.TempDir("", "TestWatchFiles")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td
		defer os.RemoveAll(tmpDir)

		// Set up dirs to watch first
		derr := mkdirAllRelative(tmpDir, expectedDirs)
		Expect(derr).NotTo(HaveOccurred())

		paths, errs := valet.WatchFiles(cancelCtx, tmpDir, valet.IsRegular, nil)

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

			for p := range paths {
				foundFiles = append(foundFiles, p)
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

	Context("when using a file predicate", func() {
		It("should find files", func() {
			Expect(len(foundFiles)).Should(Equal(len(expectedPaths)))

			for i, ep := range expectedPaths {
				a := filepath.Join(tmpDir, ep)

				fp, _ := valet.NewFilePath(a)

				Expect(foundFiles[i].Location).To(Equal(fp.Location))
				Expect(foundFiles[i].Info).ToNot(BeNil())
			}
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
	stagingDir, err := ioutil.TempDir("", "TestWatchFiles")
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
