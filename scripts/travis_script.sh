#!/bin/bash

set -e -x

. ~/miniconda/etc/profile.d/conda.sh
conda activate travis

echo "irods" | script -q -c "iinit" /dev/null
ienv
ils

which baton-do
baton-do --version

make test
