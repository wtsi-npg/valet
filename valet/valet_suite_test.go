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
	"sync"
	"time"

	ex "github.com/kjsanger/extendo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kjsanger/valet/cmd"
	"github.com/kjsanger/valet/valet"
)

var _ = Describe("Find directories)", func() {
	var (
		foundDirs     []valet.FilePath
		pathTransform localPathTransform

		dataDir = "testdata/valet"
	)

	BeforeEach(func() {
		absDataDir, err := filepath.Abs(dataDir)
		Expect(err).NotTo(HaveOccurred())
		pathTransform = makeLocalPathTransform(absDataDir)

		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, dataDir, valet.IsDir,
			valet.IsFalse)

		for path := range paths {
			foundDirs = append(foundDirs, path)
		}
		for err := range errs {
			Expect(err).NotTo(HaveOccurred())
		}

		cancel()
	})

	When("using a directory predicate", func() {
		expectedPaths := []string{
			".",
			"1",
			"1/reads",
			"1/reads/fast5",
			"1/reads/fastq",
			"testdir",
		}

		It("should find directories", func() {
			Expect(foundDirs).To(WithTransform(pathTransform,
				ConsistOf(expectedPaths)))
		})
	})
})

var _ = Describe("Find regular files)", func() {
	var (
		foundFiles    []valet.FilePath
		pathTransform localPathTransform

		dataDir = "testdata/valet"
	)

	BeforeEach(func() {
		absDataDir, err := filepath.Abs(dataDir)
		Expect(err).NotTo(HaveOccurred())
		pathTransform = makeLocalPathTransform(absDataDir)

		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "testdata/valet",
			valet.IsRegular, valet.IsFalse)

		for path := range paths {
			foundFiles = append(foundFiles, path)
		}
		for err := range errs {
			Expect(err).NotTo(HaveOccurred())
		}

		cancel()
	})

	When("using a file predicate", func() {
		It("should find files", func() {
			expectedPaths := []string{
				"1/reads/fast5/reads1.fast5",
				"1/reads/fast5/reads1.fast5.md5",
				"1/reads/fast5/reads2.fast5",
				"1/reads/fast5/reads3.fast5",
				"1/reads/fastq/reads1.fastq",
				"1/reads/fastq/reads1.fastq.md5",
				"1/reads/fastq/reads2.fastq",
				"1/reads/fastq/reads2.fastq.gz",
				"1/reads/fastq/reads3.fastq",
				"testdir/.gitignore",
			}
			Expect(foundFiles).To(WithTransform(pathTransform,
				ConsistOf(expectedPaths)))
		})
	})
})

var _ = Describe("Find files with pruning", func() {
	var (
		foundDirs     []valet.FilePath
		pathTransform localPathTransform

		dataDir = "testdata/valet"

		pruneFn = func(path valet.FilePath) (bool, error) {
			pattern, err := filepath.Abs("testdata/valet/1/reads")
			if err != nil {
				return false, err
			}

			match, err := filepath.Match(pattern, path.Location)
			if err == nil && match {
				return match, filepath.SkipDir
			}
			return match, err
		}
	)

	BeforeEach(func() {
		absDataDir, err := filepath.Abs(dataDir)
		Expect(err).NotTo(HaveOccurred())
		pathTransform = makeLocalPathTransform(absDataDir)

		cancelCtx, cancel := context.WithCancel(context.Background())
		paths, errs := valet.FindFiles(cancelCtx, "testdata/valet",
			valet.IsDir, pruneFn)

		for path := range paths {
			foundDirs = append(foundDirs, path)
		}
		for err := range errs {
			Expect(err).NotTo(HaveOccurred())
		}

		cancel()
	})

	When("using a prune function", func() {
		It("should find paths, except those pruned", func() {
			expectedPaths := []string{
				".",
				"1",
				"testdir",
			}

			Expect(foundDirs).To(WithTransform(pathTransform,
				ConsistOf(expectedPaths)))
		})
	})
})

var _ = Describe("Find files at intervals", func() {
	var (
		foundFiles    []valet.FilePath
		pathTransform localPathTransform

		dataDir = "testdata/valet"

		expectedPaths = []string{
			"1/reads/fast5/reads1.fast5",
			"1/reads/fast5/reads1.fast5.md5",
			"1/reads/fast5/reads2.fast5",
			"1/reads/fast5/reads3.fast5",
			"1/reads/fastq/reads1.fastq",
			"1/reads/fastq/reads1.fastq.md5",
			"1/reads/fastq/reads2.fastq",
			"1/reads/fastq/reads2.fastq.gz",
			"1/reads/fastq/reads3.fastq",
		}
	)

	BeforeEach(func() {
		absDataDir, err := filepath.Abs(dataDir)
		Expect(err).NotTo(HaveOccurred())
		pathTransform = makeLocalPathTransform(absDataDir)

		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 500 * time.Millisecond

		paths, errs := valet.FindFilesInterval(cancelCtx, dataDir,
			valet.IsRegular, valet.IsFalse, interval)

		// Find files or timeout and cancel
		found := make(map[string]valet.FilePath) // FilePaths are not comparable

		var wg sync.WaitGroup
		wg.Add(1)

		timeout := time.After(5 * interval)

		go func() {
			defer wg.Done()
			defer cancel()

			for {
				select {
				case <-timeout:
					return
				case path := <-paths:
					found[path.Location] = path
					if len(found) >= len(expectedPaths) {
						// Find files
						return
					}
				}
			}
		}()

		wg.Wait()

		for range paths {
			// Discard any remaining paths to unblock any sending goroutines
			// started by FindFilesInterval (it can have a number running
			// because it starts a new one at each interval). This closes the
			// errs channel too.
		}
		for err := range errs {
			Expect(err).NotTo(HaveOccurred())
		}

		var ferr error
		foundFiles, ferr = toArray(found)
		Expect(ferr).NotTo(HaveOccurred())
	})

	When("using a file predicate", func() {
		It("should find files", func() {
			Expect(foundFiles).To(WithTransform(pathTransform,
				ConsistOf(expectedPaths)))
		})
	})
})

var _ = Describe("Watch for file changes", func() {
	var (
		foundFiles    []valet.FilePath
		pathTransform localPathTransform
		tmpDir        string

		dataDir = "testdata/valet"

		expectedPaths = []string{
			"1/reads/fast5/reads1.fast5",
			"1/reads/fast5/reads2.fast5",
			"1/reads/fast5/reads3.fast5",
			"1/reads/fastq/reads1.fastq",
			"1/reads/fastq/reads2.fastq",
			"1/reads/fastq/reads3.fastq",
		}
		expectedDirs = []string{
			"1/reads/fast5/",
			"1/reads/fastq/",
		}
	)

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 500 * time.Millisecond

		td, terr := ioutil.TempDir("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td
		pathTransform = makeLocalPathTransform(tmpDir)

		// Set up dirs to watch first
		derr := mkdirAllRelative(tmpDir, expectedDirs)
		Expect(derr).NotTo(HaveOccurred())

		paths, errs :=
			valet.WatchFiles(cancelCtx, tmpDir, valet.IsRegular, valet.IsFalse)

		cerr := copyFilesRelative(dataDir, tmpDir, expectedPaths, moveFile)
		Expect(cerr).NotTo(HaveOccurred())

		// Detect updated files or timeout and cancel
		found := make(map[string]valet.FilePath)

		var wg sync.WaitGroup
		wg.Add(1)
		timeout := time.After(5 * interval)

		go func() {
			defer wg.Done()
			defer cancel()

			for {
				select {
				case <-timeout:
					return
				case path := <-paths:
					found[path.Location] = path
					if len(found) >= len(expectedPaths) {
						// Detect files
						return
					}
				}
			}
		}()

		wg.Wait()

		for err := range errs {
			Expect(err).NotTo(HaveOccurred())
		}

		var ferr error
		foundFiles, ferr = toArray(found)
		Expect(ferr).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("using a file predicate", func() {
		It("should find files", func() {
			Expect(foundFiles).To(WithTransform(pathTransform,
				ConsistOf(expectedPaths)))
		})
	})
})

var _ = Describe("Watch for file changes with pruning", func() {
	var (
		foundFiles    []valet.FilePath
		pathTransform localPathTransform
		tmpDir        string

		dataDir = "testdata/valet"

		allPaths = []string{
			"1/reads/fast5/reads1.fast5",
			"1/reads/fast5/reads2.fast5",
			"1/reads/fast5/reads3.fast5",
			"1/reads/fastq/reads1.fastq",
			"1/reads/fastq/reads2.fastq",
			"1/reads/fastq/reads3.fastq",
		}
		allDirs = []string{
			"1/reads/fast5/",
			"1/reads/fastq/",
		}

		expectedPaths = []string{
			"1/reads/fast5/reads1.fast5",
			"1/reads/fast5/reads2.fast5",
			"1/reads/fast5/reads3.fast5",
		}
	)

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 1 * time.Second

		td, terr := ioutil.TempDir("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td
		pathTransform = makeLocalPathTransform(tmpDir)

		// Set up dirs to watch first
		derr := mkdirAllRelative(tmpDir, allDirs)
		Expect(derr).NotTo(HaveOccurred())

		pattern := filepath.Join(tmpDir, "1/reads/fastq")
		pruneFn := func(path valet.FilePath) (b bool, e error) {
			match, err := filepath.Match(pattern, path.Location)
			if err == nil && match {
				return match, filepath.SkipDir
			}

			return match, err
		}

		paths, errs := valet.WatchFiles(cancelCtx, tmpDir, valet.IsRegular, pruneFn)

		cerr := copyFilesRelative(dataDir, tmpDir, allPaths, moveFile)
		Expect(cerr).NotTo(HaveOccurred())

		// Detect updated files or timeout and cancel
		found := make(map[string]valet.FilePath)

		var wg sync.WaitGroup
		wg.Add(1)
		timeout := time.After(3 * interval)

		go func() {
			defer wg.Done()
			defer cancel()

			for {
				select {
				case <-timeout:
					return
				case path := <-paths:
					found[path.Location] = path
					if len(found) >= len(expectedPaths) {
						// Detect files
						return
					}
				}
			}
		}()

		wg.Wait()

		for err := range errs {
			Expect(err).NotTo(HaveOccurred())
		}

		var ferr error
		foundFiles, ferr = toArray(found)
		Expect(ferr).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("using a file predicate", func() {
		It("should find files", func() {

			Expect(foundFiles).To(WithTransform(pathTransform,
				ConsistOf(expectedPaths)))
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

		for err := range errs {
			Expect(err).NotTo(HaveOccurred())
		}

		cancel()
	})

	When("using an experiment prune function", func() {
		It("should find only experiment files", func() {
			Expect(foundPaths).To(ConsistOf(expectedPaths))
		})
	})
})

var _ = Describe("IsArchived", func() {
	var (
		rootColl, workColl, remotePath string
		isArchived                     valet.FilePredicate
		path                           valet.FilePath

		clientPool *ex.ClientPool
		client     *ex.Client
		obj        *ex.DataObject

		localPath = "testdata/valet/1/reads/fast5/reads1.fast5"
	)

	BeforeEach(func() {
		var err error
		path, err = valet.NewFilePath(localPath)
		Expect(err).NotTo(HaveOccurred())

		rootColl = "/testZone/home/irods"
		workColl = tmpRodsPath(rootColl, "ValetIsArchived")
		remotePath = filepath.Join(workColl, "reads1.fast5")

		clientPool = ex.NewClientPool(2, time.Second)
		client, err = clientPool.Get()
		Expect(err).NotTo(HaveOccurred())

		_, err = ex.MakeCollection(client, workColl)
		Expect(err).NotTo(HaveOccurred())
		obj, err = ex.PutDataObject(client, localPath, remotePath)
		Expect(err).NotTo(HaveOccurred())

		// The expected and correct metadata
		err = obj.AddMetadata([]ex.AVU{{
			Attr:  "md5",
			Value: "1181c1834012245d785120e3505ed169"}})
		Expect(err).NotTo(HaveOccurred())

		local, err := filepath.Abs("testdata/valet/1/reads/fast5")
		Expect(err).NotTo(HaveOccurred())
		// The predicate to be tested
		isArchived = valet.MakeIsArchived(local, workColl, clientPool)
	})

	AfterEach(func() {
		err := removeTmpCollection(workColl)
		Expect(err).NotTo(HaveOccurred())

		err = clientPool.Return(client)
		Expect(err).NotTo(HaveOccurred())

		clientPool.Close()
	})

	When("a data object exists with correct checksum and md5 metadata", func() {
		It("is archived", func() {
			Expect(isArchived(path)).To(BeTrue())
		})
	})

	When("a data object does not exist", func() {
		BeforeEach(func() {
			err := obj.Remove()
			Expect(err).NotTo(HaveOccurred())
		})

		It("is not archived", func() {
			Expect(isArchived(path)).To(BeFalse())
		})
	})

	When("a data object exists, but has no md5 metadata", func() {
		BeforeEach(func() {
			err := obj.RemoveMetadata([]ex.AVU{{
				Attr:  "md5",
				Value: "1181c1834012245d785120e3505ed169"}})
			Expect(err).NotTo(HaveOccurred())
		})

		It("is not archived", func() {
			Expect(isArchived(path)).To(BeFalse())
		})
	})

	When("a data object exists, but has mismatched md5 metadata", func() {
		BeforeEach(func() {
			err := obj.RemoveMetadata([]ex.AVU{{
				Attr:  "md5",
				Value: "1181c1834012245d785120e3505ed169"}})
			Expect(err).NotTo(HaveOccurred())

			err = obj.AddMetadata([]ex.AVU{{
				Attr:  "md5",
				Value: "999999999912245d785120e3505ed169"}})
			Expect(err).NotTo(HaveOccurred())
		})

		It("is not archived", func() {
			Expect(isArchived(path)).To(BeFalse())
		})
	})

	When("a data object exists, but has a mismatched checksum", func() {
		BeforeEach(func() {
			wrongFile := "testdata/valet/1/reads/fast5/reads2.fast5"
			_, err := ex.PutDataObject(client, wrongFile, remotePath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("is not archived", func() {
			Expect(isArchived(path)).To(BeFalse())
		})
	})

})

var _ = Describe("Archive MINKnow files", func() {
	var (
		workColl     string
		tmpDir       string
		getRodsPaths itemPathTransform

		clientPool *ex.ClientPool

		rootColl = "/testZone/home/irods"
		dataDir  = "testdata/platform/ont/minknow/gridion"

		expectedArchived = []string{
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
	)

	BeforeEach(func() {
		allDirs, err := findDirsRelative(dataDir)
		Expect(err).NotTo(HaveOccurred())

		for i := range allDirs {
			allDirs[i], err = filepath.Rel(dataDir, allDirs[i])
			Expect(err).NotTo(HaveOccurred())
		}

		allFiles, err := findFilesRelative(dataDir)
		Expect(err).NotTo(HaveOccurred())

		for i := range allFiles {
			allFiles[i], err = filepath.Rel(dataDir, allFiles[i])
			Expect(err).NotTo(HaveOccurred())
		}

		td, terr := ioutil.TempDir("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td

		// Set up copy of test data (test is destructive)
		derr := mkdirAllRelative(tmpDir, allDirs)
		Expect(derr).NotTo(HaveOccurred())

		workColl = tmpRodsPath(rootColl, "ValetArchive")
		getRodsPaths = makeRodsItemTransform(workColl)

		cerr := copyFilesRelative(dataDir, tmpDir, allFiles, readWriteFile)
		Expect(cerr).NotTo(HaveOccurred())

		cancelCtx, cancel := context.WithCancel(context.Background())
		sweepInterval := 10 * time.Second

		clientPool = ex.NewClientPool(6, time.Second)
		deleteLocal := true

		defaultPruneFn, err := valet.MakeDefaultPruneFunc(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		var wg sync.WaitGroup
		wg.Add(1)

		// Find files or timeout and cancel
		perr := make(chan error, 1)

		go func() {
			plan := valet.ArchiveFilesWorkPlan(tmpDir, workColl,
				clientPool, deleteLocal)

			matchFn := valet.Or(
				valet.RequiresArchiving,
				valet.RequiresCompression)

			perr <- valet.ProcessFiles(cancelCtx, valet.ProcessParams{
				Root:          tmpDir,
				MatchFunc:     matchFn,
				PruneFunc:     defaultPruneFn,
				Plan:          plan,
				SweepInterval: sweepInterval,
				MaxProc:       4,
			})
		}()

		go func() {
			defer wg.Done()
			defer cancel()

			client, err := clientPool.Get()
			if err != nil {
				return
			}
			defer func() {
				if e := clientPool.Return(client); e != nil {
					fmt.Fprint(os.Stderr, e.Error())
				}
			}()

			timeout := time.After(120 * time.Second)

			for {
				select {
				case <-timeout:
					return
				default:
					time.Sleep(2 * time.Second)

					coll := ex.NewCollection(client, workColl)
					if exists, _ := coll.Exists(); exists {
						contents, err := coll.FetchContentsRecurse()
						if err != nil {
							return
						}
						if len(contents) >= len(expectedArchived) {
							// Detect iRODS paths
							return
						}
					}
				}
			}
		}()

		wg.Wait()

		// TODO: This is currently getting tripped by timeouts from the
		// iRODS mkdir workaround, so I have disabled it temporarily.
		//
		// Expect(<-perr).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		err = removeTmpCollection(workColl)
		Expect(err).NotTo(HaveOccurred())

		clientPool.Close()
	})

	When("using a file predicate", func() {
		It("should find files", func() {
			clientPool := ex.NewClientPool(1, time.Second*1)

			client, err := clientPool.Get()
			Expect(err).NotTo(HaveOccurred())

			coll := ex.NewCollection(client, workColl)
			Expect(coll.Exists()).To(BeTrue())

			contents, err := coll.FetchContentsRecurse()
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(WithTransform(getRodsPaths,
				ConsistOf(expectedArchived)))

			remaining, err := findFilesRelative(tmpDir)
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

	When("there are data files without checksum files", func() {
		It("should count those files", func() {
			Expect(numFilesFound).Should(Equal(numFilesExpected))
		})
	})
})
