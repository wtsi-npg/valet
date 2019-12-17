#!/bin/bash
#
# This is a wrapper script for systemd to run valet as a service. It sets up a
# Conda environment which must provide a baton-do executable >= 2.0.0 (see
# https://github.com/wtsi-npg/baton).
#
# The systemd unit may set the following environment variable to affect valet's
# behaviour. Defaults are specified in the ENVRIRONMENT block below:
#
# CONDA_ROOT : the root of a Miniconda installation
# CONDA_ENV : the name of a Conda environment
#
# HOSTNAME : the sequencing instrument hostname
# INSTRUMENT_MODEL : the sequencing instrument model (lower case)
#
# ARCHIVE_ROOT : the root collection of the iRODS data archive
# DATA_ROOT : the root data directory on the instrument
# SAFE_ROOT : a place in the instrument data filesystem that will not be archived
#
# MAX_PROC : the maximum number of processes for valet
# INTERVAL : the directory sweep interval for valet
# LOG_FILE : the combined STDOUT/STDERR log file for valet
#
# The script additionally sets TMPDIR to be in /data so that it is on the same
# filesystem as the data being processed, but also under SAFE_ROOT so that it
# doesn't get archived.

set -e

# BEGIN ENVIRONMENT
CONDA_ROOT=${CONDA_ROOT:-$HOME/miniconda}
CONDA_ENV=${CONDA_ENV:-valet}

HOSTNAME=${HOSTNAME:-$(hostname)}
INSTRUMENT_MODEL=${INSTRUMENT_MODEL:-promethion}

ARCHIVE_ROOT=${ARCHIVE_ROOT:-"/seq/ont/$INSTRUMENT_MODEL/$HOSTNAME"}
DATA_ROOT=${DATA_ROOT:-/data}
SAFE_ROOT=${SAFE_ROOT:-/data/npg}

MAX_PROC=${MAX_PROC:-10}
INTERVAL=${INTERVAL:-10m}
LOG_FILE=${LOG_FILE:-"$HOME/valet.log"}

TMPDIR=${TMPDIR:-"$SAFE_ROOT/tmp"}
# END ENVIRONMENT

source "$CONDA_ROOT/etc/profile.d/conda.sh"

set -x

conda activate "$CONDA_ENV" && \
  nice valet archive create \
  --root "$DATA_ROOT" \
  --archive-root "$ARCHIVE_ROOT" \
  --exclude "$SAFE_ROOT" \
  --exclude "$DATA_ROOT/pings" \
  --exclude "$DATA_ROOT/reports" \
  --max-proc "$MAX_PROC" \
  --interval "$INTERVAL" \
  --delete-on-archive \
  --verbose >> "$LOG_FILE" 2>&1
