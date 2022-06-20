# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - [![Unit tests](https://github.com/wtsi-npg/valet/actions/workflows/run-tests.yml/badge.svg)](https://github.com/wtsi-npg/valet/actions/workflows/run-tests.yml)

### Added

### Changed

### [1.8.0] - 2022-06-20

### Added

 - Add JSON and HTML reports to archived files
 - Add BAM indices (bai) to archived files

### Changed

 - Include archive creation script in automatic release
 - Add default ignored directories 'Install_logs' and 'persistence'

### [1.7.0] - 2022-05-19

### Added

- Remove run directories after a delay, if they are empty
- Automated GitHub release on tagging
- Add iRODS 4.2.11 to GitHub Actions workflow

### Changed

- Bump github.com/onsi/ginkgo from 1.16.5 to 2.1.4
- Bump github.com/onsi/gomega from 1.17.0 to 1.19.0
- Bump github.com/stretchr/testify from 1.7.0 to 1.7.1
- Bump github.com/spf13/cobra from 1.2.1 to 1.4.0

## [1.6.0] - 2022-01-10

### Added

- Support for adding iRODS metadata for PromethION-24 runs.
- Support for BAM files.
- support for BED files (compressed).
- iRODS 4.2.10 support.

### Changed

- Build with Go.1.17

- Bump github.com/rs/zerolog from 1.21.0 to 1.26.1
- Bump github.com/onsi/ginkgo from 1.16.2 to 1.16.5
- Bump github.com/onsi/gomega from 1.11.0 to 1.17.0
- Bump github.com/spf13/cobra from 0.0.7 to 1.3.0

### Removed

- Travis CI configuration.
- PromethION beta report parsing tests and unused test data.

## [1.5.0] - 2021-04-15

### Added

- Github Actions test workflow.
- Improve recognition of files to be processed and archived.
- Support for TSV files.
- Support for compressing CSV files.
- Add all ONT-specific paths to the exclusion list.
- Add the /data/laboratory WSI-specific path to the exclusion list.
- iRODS 4.2.8 support.

### Removed

- RODS 4.1.12 support.

### Changed

- Relocate repository from github.com/kjsanger to github.com/wtsi-npg

- Bump github.com/wtsi-npg/extendo/v2 from 2.2.0 to 2.4.0
- Bump github.com/wtsi-npg/logshim from to 1.1.0 to 1.3.0
- Bump github.com/wtsi-npg/logshim-zerolog from 1.1.0 to 1.3.0
- Bump github.com/klauspost/pgzip from 1.2.1 to 1.2.5
- Bump github.com/onsi/ginkgo from 1.12.3 to 1.16.1
- Bump github.com/onsi/gomega from 1.10.1 to 1.11.0
- Bump github.com/rs/zerolog from 1.19.1 to 1.21.0
- Bump github.com/spf13/cobra from 0.0.5 to 0.0.7

## [1.4.0] - 2020-09-14

### Added

- PromethION support for enhanced metadata.
- iRODS 4.2.7 clients to the test matrix.
- "ont" namespace to iRODS metadata attributes.

### Changed

- Bump github.com/wtsi-npg/extendo from 2.1.0 to 2.2.0

## [1.3.0] - 2020-06-02

### Added

- experiment_name and instrument_slot to report metadata.

### Changed

- Bump github.com/wtsi-npg/extendo from 2.0.0 to 2.1.0
- Bump github.com/stretchr/testify from 1.5.1 to 1.6.0
- Bump github.com/onsi/ginkgo from 1.12.0 to 1.12.2

## [1.2.0] - 2020-02-28

### Added

- A systemd unit template and wrapper script.
- A String method to WorkPlan.
- Logging immediately before and after a work function is called.

### Changed

- Improve verbose level logging for consistency and information.
- Add to default paths to ignore in the data root.

- Bump github.com/wtsi-npg/extendo from 1.1.0 to 2.0.0
- Bump github.com/onsi/ginkgo from 1.10.3 to 1.12.0
- Bump github.com/onsi/gomega from 1.7.1 to 1.9.0

### Fixed

- Hang or failure to exit cleanly when encountering errors, such as
  unreadable directories.
- Hang on cancel when data root does not exist.
- Exclude TMPDIR from archiving.
- A --version CLI option.

## [1.1.0] - 2019-12-05

### Added

- Checksum creation improvements using tee'ing to avoid re-reading data.
- Make the size of the client pool equal to the number of threads.
- Gzip compression of fastq and txt files.
- Enhancements to FilePath for handling compressed files.
- Find paths immediately, then at intervals.

## Changed

- Use pgzip implementation.

- Bump github.com/wtsi-npg/extendo from 1.0.0 to 1.1.0
- Bump github.com/onsi/gomega from 1.5.0 to 1.7.1
- Bump github.com/onsi/ginkgo from 1.8.0 to 1.10.3
- Bump github.com/rs/zerolog from 1.14.3 to 1.17.2
- Bump github.com/stretchr/testify from 1.3.0 to 1.4.0

### Fixed

- Check the error returned by ProcessFiles.
- Improvements to error handling in ProcessFiles.
- Allow cancellation of work to cause ProcessFiles to return.
- Compress files into a temporary location.


## [1.0.0] - 2019-10-14

### Added

- Limit the absolute maximum number of threads to 12.
- A --delete-on-archive CLI option to valet archive.
- A valet archive command to move files to iRODS using extendo.
- Support for dispatching to multiple different work functions.
- A --dry-run CLI option to valet checksum.
- Watch and sweep tree pruning capability.
- Signal handler to call the cancel function on SIGINT and SIGTERM.
