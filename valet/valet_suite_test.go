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
	"time"

	ex "github.com/kjsanger/extendo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kjsanger/valet/cmd"
	"github.com/kjsanger/valet/valet"
)

var _ = Describe("Find directories)", func() {
	var (
		foundDirs []valet.FilePath

		expectedPaths = []string{
			"testdata/valet",
			"testdata/valet/1",
			"testdata/valet/1/reads",
			"testdata/valet/1/reads/fast5",
			"testdata/valet/1/reads/fastq",
			"testdata/valet/testdir",
		}
	)

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "testdata/valet",
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
	var (
		foundFiles []valet.FilePath

		expectedPaths = []string{
			"testdata/valet/1/reads/fast5/reads1.fast5",
			"testdata/valet/1/reads/fast5/reads1.fast5.md5",
			"testdata/valet/1/reads/fast5/reads2.fast5",
			"testdata/valet/1/reads/fast5/reads3.fast5",
			"testdata/valet/1/reads/fastq/reads1.fastq",
			"testdata/valet/1/reads/fastq/reads1.fastq.md5",
			"testdata/valet/1/reads/fastq/reads2.fastq",
			"testdata/valet/1/reads/fastq/reads2.fastq.gz",
			"testdata/valet/1/reads/fastq/reads3.fastq",
			"testdata/valet/testdir/.gitignore",
		}
	)

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "testdata/valet",
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
	var (
		expectedPaths = []string{
			"testdata/valet",
			"testdata/valet/1",
			"testdata/valet/testdir",
		}

		pruneFn = func(pf valet.FilePath) (bool, error) {
			pattern, err := filepath.Abs("testdata/valet/1/reads")
			if err != nil {
				return false, err
			}

			match, err := filepath.Match(pattern, pf.Location)
			if err == nil && match {
				return match, filepath.SkipDir
			}
			return match, err
		}

		foundDirs []valet.FilePath
	)

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "testdata/valet",
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

	Context("Using a prune function", func() {
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
	var (
		expectedPaths = []string{
			"testdata/valet/1/reads/fast5/reads1.fast5",
			"testdata/valet/1/reads/fast5/reads1.fast5.md5",
			"testdata/valet/1/reads/fast5/reads2.fast5",
			"testdata/valet/1/reads/fast5/reads3.fast5",
			"testdata/valet/1/reads/fastq/reads1.fastq",
			"testdata/valet/1/reads/fastq/reads1.fastq.md5",
			"testdata/valet/1/reads/fastq/reads2.fastq",
			"testdata/valet/1/reads/fastq/reads2.fastq.gz",
			"testdata/valet/1/reads/fastq/reads3.fastq",
		}

		foundFiles = map[string]bool{}
	)

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 500 * time.Millisecond

		paths, errs := valet.FindFilesInterval(cancelCtx, "testdata/valet",
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
	var (
		tmpDir     string
		foundFiles = map[string]bool{}

		expectedPaths = []string{
			"testdata/valet/1/reads/fast5/reads1.fast5",
			"testdata/valet/1/reads/fast5/reads2.fast5",
			"testdata/valet/1/reads/fast5/reads3.fast5",
			"testdata/valet/1/reads/fastq/reads1.fastq",
			"testdata/valet/1/reads/fastq/reads2.fastq",
			"testdata/valet/1/reads/fastq/reads3.fastq",
		}
		expectedDirs = []string{
			"testdata/valet/1/reads/fast5/",
			"testdata/valet/1/reads/fastq/",
		}
	)

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
	var (
		tmpDir     string
		foundFiles = map[string]bool{}

		allPaths = []string{
			"testdata/valet/1/reads/fast5/reads1.fast5",
			"testdata/valet/1/reads/fast5/reads2.fast5",
			"testdata/valet/1/reads/fast5/reads3.fast5",
			"testdata/valet/1/reads/fastq/reads1.fastq",
			"testdata/valet/1/reads/fastq/reads2.fastq",
			"testdata/valet/1/reads/fastq/reads3.fastq",
		}
		allDirs = []string{
			"testdata/valet/1/reads/fast5/",
			"testdata/valet/1/reads/fastq/",
		}

		expectedPaths = allPaths[:4]
	)

	pruneFn := func(pf valet.FilePath) (bool, error) {
		pattern, err := filepath.Abs("testdata/valet/1/reads/fastq")
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

var _ = Describe("Find MinKNOW files", func() {
	var (
		foundPaths []string

		dataDir = "testdata/platform/ont/minknow/gridion"

		expectedPaths = []string{
			".",
			"66",
			"66/DN585561I_A1",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_pass",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_fail",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_pass",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_fail",
		}
	)

	BeforeEach(func() {
		patterns, err := valet.DefaultIgnorePatterns(dataDir)
		Expect(err).NotTo(HaveOccurred())

		pruneFn, err := valet.MakeGlobPruneFunc(patterns)
		Expect(err).NotTo(HaveOccurred())

		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, dataDir, valet.IsDir, pruneFn)

		absData, err := filepath.Abs(dataDir)
		Expect(err).NotTo(HaveOccurred())

		for path := range paths {
			relPath, err := filepath.Rel(absData, path.Location)
			Expect(err).NotTo(HaveOccurred())
			foundPaths = append(foundPaths, relPath)
		}

		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		default:
		}

		cancel()
	})

	Context("Using an experiment prune function", func() {
		It("should find only experiment files", func() {
			Expect(foundPaths).To(ConsistOf(expectedPaths))
		})
	})
})

var _ = Describe("Archive MINKnow files", func() {
	var (
		allFiles []string
		allDirs  []string

		tmpDir, tmpDataDir string
		rootColl, workColl string

		dataDir = "testdata/platform/ont/minknow/gridion"

		getRodsPaths itemPathTransform
	)

	BeforeEach(func() {
		ad, err := findDirsRelative(dataDir)
		allDirs = ad
		Expect(err).NotTo(HaveOccurred())

		af, err := findFilesRelative(dataDir)
		allFiles = af
		Expect(err).NotTo(HaveOccurred())

		td, terr := ioutil.TempDir("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td
		tmpDataDir = filepath.Join(tmpDir, "testdata/platform/ont/minknow/gridion")

		rootColl = "/testZone/home/irods"
		workColl = tmpRodsPath(rootColl, "ValetArchive")

		// Set up copy of test data (test is destructive)
		derr := mkdirAllRelative(tmpDir, allDirs)
		Expect(derr).NotTo(HaveOccurred())

		getRodsPaths = makeRodsItemTransform(workColl)

		cerr := copyFilesRelative(tmpDir, allFiles, readWriteFile)
		Expect(cerr).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		err = removeTmpCollection(workColl)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when using a file predicate", func() {
		It("should find files", func() {
			cancelCtx, cancel := context.WithCancel(context.Background())
			sweepInterval := 10 * time.Second

			clientPool := ex.NewClientPool(6, time.Second*1)
			deleteLocal := true

			defaultPruneFn, err := valet.MakeDefaultPruneFunc(tmpDataDir)
			Expect(err).NotTo(HaveOccurred())

			expectedArchived := []string{
				".",
				"66",
				"66/DN585561I_A1",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_fail",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_pass",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_fail",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_pass",

				// Ancillary files
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/" +
					"GXB02004_20190904_151413_FAL01979_gridion_sequencing_run_" +
					"DN585561I_A1_sequencing_summary.txt.gz",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/" +
					"final_summary.txt.gz",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/duty_time.csv",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/report.md",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/report.pdf",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/throughput.csv",

				// Fast5 fail
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_fail/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.fast5",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_fail/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.fast5",
				// Fast5 pass
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_pass/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_0.fast5",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_pass/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.fast5",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_pass/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.fast5",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_pass/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_3.fast5",
				// Fastq fail
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_fail/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.fastq.gz",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_fail/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.fastq.gz",
				// Fastq pass
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_pass/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.fastq.gz",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_pass/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.fastq.gz",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_pass/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_3.fastq.gz",
				"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_pass/" +
					"FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_4.fastq.gz",
			}

			// Find files or timeout and cancel
			done := make(chan bool, 2)

			go func() {
				plan := valet.ArchiveFilesWorkPlan(tmpDataDir, workColl,
					clientPool, deleteLocal)

				matchFn := valet.Or(
					valet.RequiresArchiving,
					valet.RequiresCompression)

				valet.ProcessFiles(cancelCtx, valet.ProcessParams{
					Root:          tmpDataDir,
					MatchFunc:     matchFn,
					PruneFunc:     defaultPruneFn,
					Plan:          plan,
					SweepInterval: sweepInterval,
					MaxProc:       4,
				})
			}()

			go func() {
				defer cancel()

				timer := time.NewTimer(120 * time.Second)
				<-timer.C
				done <- true // Timeout
			}()

			go func() {
				defer cancel()

				client, err := clientPool.Get()
				if err != nil {
					return
				}

				for {
					time.Sleep(2 * time.Second)

					coll := ex.NewCollection(client, workColl)
					if exists, _ := coll.Exists(); exists {
						contents, err := coll.FetchContentsRecurse()
						if err != nil {
							return
						}

						if len(contents) >= len(expectedArchived) {
							done <- true // Detect iRODS paths
							return
						}
					}
				}
			}()

			<-done

			client, err := clientPool.Get()
			Expect(err).NotTo(HaveOccurred())

			coll := ex.NewCollection(client, workColl)
			Expect(coll.Exists()).To(BeTrue())

			contents, err := coll.FetchContentsRecurse()
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(WithTransform(getRodsPaths,
				ConsistOf(expectedArchived)))

			remaining, err := findFilesRelative(tmpDataDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(BeEmpty())
		})
	})
})

var _ = Describe("Count files without a checksum", func() {
	var (
		numFilesFound    uint64
		numFilesExpected uint64 = 3 // Only fast5 and fastq.gz
	)

	BeforeEach(func() {
		n, err := cmd.CountFilesWithoutChecksum("testdata/valet", []string{})
		Expect(err).NotTo(HaveOccurred())
		numFilesFound = n
	})

	Context("when there are data files without checksum files", func() {
		It("should count those files", func() {
			Expect(numFilesFound).Should(Equal(numFilesExpected))
		})
	})
})
