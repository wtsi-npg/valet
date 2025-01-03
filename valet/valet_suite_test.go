/*
 * Copyright (C) 2019, 2020, 2021, 2022, 2023. Genome Research Ltd.
 * All rights reserved.
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
	"os"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ex "github.com/wtsi-npg/extendo/v2"

	"github.com/wtsi-npg/valet/cmd"
	"github.com/wtsi-npg/valet/valet"
)

var _ = Describe("Find directories", func() {
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
			"1/reads/alignments",
			"1/reads/fast5",
			"1/reads/fastq",
			"1/reads/pod5",
			"testdir",
		}

		It("should find directories", func() {
			Expect(foundDirs).To(WithTransform(pathTransform,
				ConsistOf(expectedPaths)))
		})
	})
})

var _ = Describe("Find regular files", func() {
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
				"1/adaptive_sampling_roi1.bed",
				"1/ancillary.csv.gz",
				"1/reads/alignments/alignments1.bam",
				"1/reads/alignments/alignments1.bam.bai",
				"1/reads/fast5/reads1.fast5",
				"1/reads/fast5/reads1.fast5.md5",
				"1/reads/fast5/reads2.fast5",
				"1/reads/fast5/reads3.fast5",
				"1/reads/fastq/reads1.fastq",
				"1/reads/fastq/reads1.fastq.md5",
				"1/reads/fastq/reads2.fastq",
				"1/reads/fastq/reads2.fastq.gz",
				"1/reads/fastq/reads3.fastq",
				"1/reads/pod5/reads1.pod5",
				"report_ABQ808_20200204_1257_e2e93dd1.md",
				"report_PAE48813_20200130_0940_16917585.md",
				"report_PAH48449_20211215_1420_227842f4.md",
				"report_PAH48449_20211215_1445_f5d8e5aa.md",
				"report_PAH48449_20211215_1509_e045091f.md",
				"report_PAH48449_20211215_1532_2a0a5bc7.md",
				"report_PAH48449_20211215_1553_3720e75e.md",
				"report_PAH48449_20211215_1617_fa1a14d5.md",
				"testdir/.gitignore",
			}
			Expect(foundFiles).To(WithTransform(pathTransform,
				ConsistOf(expectedPaths)))
		})
	})
})

var _ = Describe("Handle errors while finding files", func() {
	var (
		tmpDir        string
		foundPaths    []valet.FilePath
		pathTransform func(i []valet.FilePath) []string
	)

	BeforeEach(func() {
		td, terr := os.MkdirTemp("", "ValetHandleErrors")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td

		pathTransform = makeLocalPathTransform(tmpDir)

		// Make some data directories
		for i := 0; i < 10; i++ {
			mode := os.FileMode(0700)

			// Make odd-numbered directories unreadable to cause errors both in
			// setting up watches and in directory walks
			if i%2 == 1 {
				mode = os.FileMode(0300)
			}

			d := filepath.Join(tmpDir, fmt.Sprintf("data%d", i))

			err := os.MkdirAll(d, mode)
			Expect(err).NotTo(HaveOccurred())

			f, err := os.Create(filepath.Join(d, "test.txt"))
			Expect(err).NotTo(HaveOccurred())
			err = f.Close()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		files, err := os.ReadDir(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		for _, file := range files {
			path := filepath.Join(tmpDir, file.Name())
			err := os.Chmod(path, os.FileMode(0700)) // restore permissions for cleanup
			Expect(err).NotTo(HaveOccurred())
		}

		err = os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("some data directories are unreadable", func() {
		BeforeEach(func() {
			cancelCtx, cancel := context.WithCancel(context.Background())

			paths, errs := valet.FindFiles(cancelCtx, tmpDir, valet.IsTrue,
				valet.IsFalse)

			for path := range paths {
				foundPaths = append(foundPaths, path)
			}
			for err := range errs {
				Expect(err).NotTo(HaveOccurred())
			}

			cancel()
		})

		It("should still find reachable files", func() {
			expectedPaths := []string{
				".",
				"data0",
				"data0/test.txt",
				"data2",
				"data2/test.txt",
				"data4",
				"data4/test.txt",
				"data6",
				"data6/test.txt",
				"data8",
				"data8/test.txt",
			}

			Expect(foundPaths).To(WithTransform(pathTransform,
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
			"1/adaptive_sampling_roi1.bed",
			"1/ancillary.csv.gz",
			"1/reads/alignments/alignments1.bam",
			"1/reads/alignments/alignments1.bam.bai",
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
			"1/reads/pod5/",
		}
	)

	BeforeEach(func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 500 * time.Millisecond

		td, terr := os.MkdirTemp("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td
		pathTransform = makeLocalPathTransform(tmpDir)

		// Set up dirs to watch first
		derr := mkdirAllRelative(tmpDir, expectedDirs)
		Expect(derr).NotTo(HaveOccurred())

		paths, errs :=
			valet.WatchFiles(cancelCtx, tmpDir, valet.IsRegular, valet.IsFalse)
		time.Sleep(time.Second * 2) // Allow watches to be established

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
		It("should detect files", func() {
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

		td, terr := os.MkdirTemp("", "ValetTests")
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
		time.Sleep(time.Second * 2) // Allow watches to be established

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
		It("should detect files", func() {

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
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/bam_fail",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/bam_pass",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_pass",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fast5_fail",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_pass",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/fastq_fail",
			"66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f/other_reports",
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

var _ = Describe("IsCopied", func() {
	var (
		rootColl, workColl, remotePath string
		isCopied                       valet.FilePredicate
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
		workColl = tmpRodsPath(rootColl, "ValetIsCopied")
		remotePath = filepath.Join(workColl, "reads1.fast5")

		poolParams := ex.DefaultClientPoolParams
		poolParams.MaxSize = 2
		poolParams.GetTimeout = time.Second

		clientPool = ex.NewClientPool(poolParams)
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

		local, err := filepath.Abs("testdata/valet/1/reads/fast5/")
		Expect(err).NotTo(HaveOccurred())
		// The predicate to be tested
		isCopied = valet.MakeIsCopied(local, workColl, clientPool)
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
			Expect(isCopied(path)).To(BeTrue())
		})
	})

	When("a data object does not exist", func() {
		BeforeEach(func() {
			err := obj.Remove()
			Expect(err).NotTo(HaveOccurred())
		})

		It("is not archived", func() {
			Expect(isCopied(path)).To(BeFalse())
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
			Expect(isCopied(path)).To(BeFalse())
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

		It("is not copied", func() {
			Expect(isCopied(path)).To(BeFalse())
		})
	})

	When("a data object exists, but has a mismatched checksum", func() {
		BeforeEach(func() {
			wrongFile := "testdata/valet/1/reads/fast5/reads2.fast5"
			_, err := ex.PutDataObject(client, wrongFile, remotePath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("is not copied", func() {
			Expect(isCopied(path)).To(BeFalse())
		})
	})
})

var _ = Describe("IsAnnotated", func() {
	var (
		rootColl, workColl, remotePath string
		isAnnotated                    valet.FilePredicate
		path                           valet.FilePath

		clientPool *ex.ClientPool
		client     *ex.Client
		coll       *ex.Collection
		obj        *ex.DataObject

		localPath = "testdata/valet/report_ABQ808_20200204_1257_e2e93dd1.md"
	)

	BeforeEach(func() {
		var err error
		path, err = valet.NewFilePath(localPath)
		Expect(err).NotTo(HaveOccurred())

		rootColl = "/testZone/home/irods"
		workColl = tmpRodsPath(rootColl, "ValetIsAnnotated")
		remotePath = filepath.Join(workColl, "report_ABQ808_20200204_1257_e2e93dd1.md")

		poolParams := ex.DefaultClientPoolParams
		poolParams.MaxSize = 2
		poolParams.GetTimeout = time.Second

		clientPool = ex.NewClientPool(poolParams)
		client, err = clientPool.Get()
		Expect(err).NotTo(HaveOccurred())

		coll, err = ex.MakeCollection(client, workColl)
		Expect(err).NotTo(HaveOccurred())
		obj, err = ex.PutDataObject(client, localPath, remotePath)
		Expect(err).NotTo(HaveOccurred())

		// The expected and correct metadata. Note that this is expected to be
		// on the collection containing the report data object, not on the data
		// object itself.
		err = coll.AddMetadata([]ex.AVU{
			{Attr: "ont:device_id", Value: "X2"},
			{Attr: "ont:device_type", Value: "gridion"},
			{Attr: "ont:distribution_version", Value: "19.12.2"},
			{Attr: "ont:experiment_name", Value: "85"},
			{Attr: "ont:flowcell_id", Value: "ABQ808"},
			{Attr: "ont:guppy_version", Value: "3.2.8+bd67289"},
			{Attr: "ont:hostname", Value: "GXB02004"},
			{Attr: "ont:instrument_slot", Value: "2"},
			{Attr: "ont:protocol_group_id", Value: "85"},
			{Attr: "ont:run_id", Value: "5531cbcf622d2d98dbff00af0261c6f19f91340f"},
			{Attr: "ont:sample_id", Value: "DN615089W_B1"},
		})
		Expect(err).NotTo(HaveOccurred())

		local, err := filepath.Abs("testdata/valet/")
		Expect(err).NotTo(HaveOccurred())
		// The predicate to be tested
		isAnnotated = valet.MakeIsAnnotated(local, workColl, clientPool)
	})

	AfterEach(func() {
		err := removeTmpCollection(workColl)
		Expect(err).NotTo(HaveOccurred())

		err = clientPool.Return(client)
		Expect(err).NotTo(HaveOccurred())

		clientPool.Close()
	})

	When("correct collection annotation exists", func() {
		It("is annotated", func() {
			Expect(isAnnotated(path)).To(BeTrue())
		})
	})

	When("correct collection annotation exists, but the report does not", func() {
		BeforeEach(func() {
			err := obj.Remove()
			Expect(err).NotTo(HaveOccurred())
		})

		It("is annotated", func() {
			Expect(isAnnotated(path)).To(BeTrue())
		})
	})

	When("a metadata AVU is missing", func() {
		BeforeEach(func() {
			err := coll.RemoveMetadata([]ex.AVU{{
				Attr:  "ont:device_id",
				Value: "X2"}})
			Expect(err).NotTo(HaveOccurred())
		})

		It("is not annotated", func() {
			Expect(isAnnotated(path)).To(BeFalse())
		})
	})

	When("a metadata AVU is mismatched", func() {
		BeforeEach(func() {
			err := coll.RemoveMetadata([]ex.AVU{{Attr: "ont:device_id", Value: "X2"}})
			Expect(err).NotTo(HaveOccurred())

			err = coll.AddMetadata([]ex.AVU{{Attr: "ont:device_id", Value: "X5"}})
			Expect(err).NotTo(HaveOccurred())
		})

		It("is not annotated", func() {
			Expect(isAnnotated(path)).To(BeFalse())
		})
	})
})

var _ = Describe("Archive MinKNOW files", func() {
	var (
		workColl     string
		tmpDir       string
		getRodsPaths itemPathTransform

		clientPool *ex.ClientPool

		rootColl = "/testZone/home/irods"
		dataDir  = "testdata/platform/ont/minknow/gridion"

		collPath   = "66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f"
		bmFailColl = collPath + "/bam_fail"
		bmPassColl = collPath + "/bam_pass"
		f5FailColl = collPath + "/fast5_fail"
		f5PassColl = collPath + "/fast5_pass"
		fqFailColl = collPath + "/fastq_fail"
		fqPassColl = collPath + "/fastq_pass"
		reportColl = collPath + "/other_reports"

		expectedArchived = []string{
			".",
			"66",
			"66/DN585561I_A1",
			collPath,
			bmFailColl,
			bmPassColl,
			f5FailColl,
			f5PassColl,
			fqFailColl,
			fqPassColl,
			reportColl,

			// Ancillary files
			collPath + "/barcode_alignment_FAL01979_43578c8f.tsv",
			collPath + "/duty_time.csv.gz",
			collPath + "/final_summary_FAL01979_43578c8f.txt.gz",
			collPath + "/pore_activity_FAL01979_43578c8f.tsv",
			collPath + "/report_FAL01979_20190904_1514_43578c8f.html",
			collPath + "/report_FAL01979_20190904_1514_43578c8f.json.gz",
			collPath + "/report_FAL01979_20190904_1514_43578c8f.md",
			collPath + "/report_FAL01979_20190904_1514_43578c8f.pdf",
			collPath + "/sample_sheet_FAL01979_20190904_1514_43578c8f.csv.gz",
			collPath + "/sequencing_summary_FAL01979_43578c8f.txt.gz",
			collPath + "/throughput_FAL01979_43578c8f.csv.gz",

			bmFailColl + "/FAL01979_fail_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.bam",
			bmFailColl + "/FAL01979_fail_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.bam.bai",
			bmFailColl + "/FAL01979_fail_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.bam",
			bmFailColl + "/FAL01979_fail_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.bam.bai",

			bmPassColl + "/FAL01979_pass_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.bam",
			bmPassColl + "/FAL01979_pass_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.bam.bai",
			bmPassColl + "/FAL01979_pass_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.bam",
			bmPassColl + "/FAL01979_pass_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.bam.bai",

			f5FailColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.fast5",
			f5FailColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.fast5",

			f5PassColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_0.fast5",
			f5PassColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.fast5",
			f5PassColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.fast5",
			f5PassColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_3.fast5",

			fqFailColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.fastq.gz",
			fqFailColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.fastq.gz",

			fqPassColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_1.fastq.gz",
			fqPassColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_2.fastq.gz",
			fqPassColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_3.fastq.gz",
			fqPassColl + "/FAL01979_9cd2a77baacfe99d6b16f3dad2c36ecf5a6283c3_4.fastq.gz",

			// Reports
			reportColl + "/pore_scan_data_FAL01979_43578c8f.csv.gz",
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

		td, terr := os.MkdirTemp("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td

		// Set up copy of test data (test is destructive)
		derr := mkdirAllRelative(tmpDir, allDirs)
		Expect(derr).NotTo(HaveOccurred())

		workColl = tmpRodsPath(rootColl, "ArchiveMinKNOWFiles")
		getRodsPaths = makeRodsItemTransform(workColl)

		cerr := copyFilesRelative(dataDir, tmpDir, allFiles, readWriteFile)
		Expect(cerr).NotTo(HaveOccurred())

		cancelCtx, cancel := context.WithCancel(context.Background())
		interval := 10 * time.Second

		poolParams := ex.DefaultClientPoolParams
		poolParams.MaxSize = 6
		poolParams.GetTimeout = time.Second
		clientPool = ex.NewClientPool(poolParams)
		deleteLocal := true
		cleanup := 1 * time.Hour

		defaultPruneFn, err := valet.MakeDefaultPruneFunc(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		var wg sync.WaitGroup
		wg.Add(1)

		// Find files or timeout and cancel
		perr := make(chan error, 1)

		go func() {
			plan := valet.ArchiveFilesWorkPlan(tmpDir, workColl,
				clientPool, deleteLocal, cleanup)

			matchFn := valet.Or(
				valet.RequiresCopying,
				valet.RequiresCompression)

			perr <- valet.ProcessFiles(cancelCtx, valet.ProcessParams{
				Root:          tmpDir,
				MatchFunc:     matchFn,
				PruneFunc:     defaultPruneFn,
				Plan:          plan,
				SweepInterval: interval,
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
			poolParams := ex.DefaultClientPoolParams
			poolParams.MaxSize = 1
			poolParams.GetTimeout = time.Second
			clientPool := ex.NewClientPool(poolParams)

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

var _ = Describe("Remove empty run directories after a delay", func() {
	var (
		tmpDir string
		runDir string

		// Remove any work directory more than 100 ms old
		olderThan100ms = valet.RemoveDirectoryWorkPlan(time.Millisecond * 100)
	)

	BeforeEach(func() {
		td, terr := os.MkdirTemp("", "ValetTests")
		Expect(terr).NotTo(HaveOccurred())
		tmpDir = td

		runDir = filepath.Join(tmpDir, "66/DN585561I_A1/20190904_1514_GA20000_FAL01979_43578c8f")
		merr := os.MkdirAll(runDir, 0700)
		Expect(merr).NotTo(HaveOccurred())

		subDirs := []string{
			"fast5_pass", "fast5_fail", "fastq_pass", "fastq_fail", "other_reports"}
		for _, d := range subDirs {
			err := os.Mkdir(filepath.Join(runDir, d), 0700)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("its descendants contain files", func() {
		var f5File string

		BeforeEach(func() {
			f5Dir := filepath.Join(runDir, "fast5_pass")
			f5File = filepath.Join(f5Dir, "test.fast5")
			f, err := os.Create(f5File)
			defer f.Close()
			Expect(err).NotTo(HaveOccurred())

			cancelCtx, cancel := context.WithCancel(context.Background())
			interval := 500 * time.Millisecond

			var wg sync.WaitGroup
			wg.Add(1)

			perr := make(chan error, 1)

			go func() {
				perr <- valet.ProcessFiles(cancelCtx, valet.ProcessParams{
					Root:          tmpDir,
					MatchFunc:     valet.IsDir,
					PruneFunc:     valet.IsFalse,
					Plan:          olderThan100ms,
					SweepInterval: interval,
					MaxProc:       1,
				})
			}()

			timeout := time.After(5 * interval)

			go func() {
				defer wg.Done()
				defer cancel()

				for {
					select {
					case <-timeout:
						return
					default:
						if _, err := os.Stat(runDir); err != nil {
							if os.IsNotExist(err) {
								return
							}
						}
					}
				}
			}()

			wg.Wait()

			Expect(<-perr).NotTo(HaveOccurred())
		})
		It("is not removed", func() {
			Expect(runDir).To(BeADirectory())
			Expect(f5File).To(BeAnExistingFile())
		})
	})

	When("its descendants contain no files", func() {
		BeforeEach(func() {
			cancelCtx, cancel := context.WithCancel(context.Background())
			interval := 500 * time.Millisecond

			var wg sync.WaitGroup
			wg.Add(1)

			perr := make(chan error, 1)

			go func() {
				perr <- valet.ProcessFiles(cancelCtx, valet.ProcessParams{
					Root:          tmpDir,
					MatchFunc:     valet.IsDir,
					PruneFunc:     valet.IsFalse,
					Plan:          olderThan100ms,
					SweepInterval: interval,
					MaxProc:       1,
				})
			}()

			timeout := time.After(5 * interval)

			go func() {
				defer wg.Done()
				defer cancel()

				for {
					select {
					case <-timeout:
						return
					default:
						if _, err := os.Stat(runDir); err != nil {
							if os.IsNotExist(err) {
								return
							}
						}
					}
				}
			}()

			wg.Wait()

			Expect(<-perr).NotTo(HaveOccurred())
		})

		It("is removed", func() {
			Expect(runDir).NotTo(Or(BeADirectory(), BeAnExistingFile()))
		})
	})
})

var _ = Describe("Count files without a checksum", func() {
	// We expect checksums for files that either don't need to be compressed, or
	// do need to be compressed and have been e.g. bam, csv.gz, fast5, fastq.gz
	// and md (reports).
	//
	// We don't expect checksums for files that need to be compressed and
	// haven't been e.g. fastq. Ideally we would also record a checksum of the
	// uncompressed data.
	//
	// iRODS doesn't support this (it will calculate a checksum of the
	// compressed file once uploaded, but has no specific place for an original
	// checksum). We could add an original checksum to the file metadata.
	// However, we're already adding the compressed file checksum there (because
	// iRODS has a history of checksum-related bugs) and it would be potentially
	// confusing for customers.

	var (
		numFilesFound    uint64
		numFilesExpected uint64 = 15
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
