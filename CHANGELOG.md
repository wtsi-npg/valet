# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - [![Unit tests](https://github.com/wtsi-npg/valet/actions/workflows/run-tests.yml/badge.svg)](https://github.com/wtsi-npg/valet/actions/workflows/run-tests.yml)

### Added

### Changed

## [1.5.0] - 2021-04-15

### Added

- Add Github Actions test workflow
- Improve recognition of files to be processed and archived
- Add support for TSV files.
- Add support for compressing CSV files.
- Add new all ONT-specific paths to the exclusion list.
- Add the /data/laboratory WSI-specific path to the exclusion list.
- Add iRODS 4.2.8 support

### Removed

- Remove iRODS 4.1.12 support

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
- Use pgzip implementation.
- Gzip compression of fastq and txt files.
- Enhancements to FilePath for handling compressed files.
- Find paths immediately, then at intervals.

## Changed

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
