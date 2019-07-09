# valet

## Overview

`valet` is a utility for performing small, but important data management tasks
automatically. Once started, `valet` will continue working until interrupted
by SIGINT or SIGTERM, when it will stop gracefully.

### Tasks

- Creating up-to-date checksum files

  - Directory hierarchy styles supported
    
    - Any
  
  - File patterns supported
  
    - *.fast5$
    - *.fastq$

  - Checksum file patterns supported
  
    - (data file name).md5

`valet` will monitor a directory hierarchy and locate data files within it that
have no accompanying checksum file, or have a checksum file that is stale.
`valet` will then calculate the checksum and create or update the checksum file.



