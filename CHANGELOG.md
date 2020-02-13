# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - [![Build Status](https://travis-ci.org/kjsanger/valet.svg?branch=devel)](https://travis-ci.org/kjsanger/valet)

### Added

- A systemd unit template and wrapper script.
- A String() method for WorkPlan.   
- Logging immediately before and after a work function is called.

### Changed

- Improved verbose level logging for consistency and information.
- Added to default paths to ignore in the data root.

- Bump github.com/onsi/ginkgo from 1.10.3 to 1.12.0
- Bump github.com/onsi/gomega from 1.7.1 to 1.9.0

### Fixed

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

- Bump github.com/kjsanger/extendo from 1.0.0 to 1.1.0
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
