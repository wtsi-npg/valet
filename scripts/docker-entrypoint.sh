#!/bin/bash

set -e

if type "iinit" > /dev/null 2>&1 ; then
    if [ -n "$IRODS_PASSWORD" ]; then
        echo "Running iinit with the password provided in IRODS_PASSWORD environment variable"
        echo "$IRODS_PASSWORD" | iinit
    else
        echo "No password provided in the IRODS_PASSWORD environment variable. Expecting an iRODS auth file in a mounted volume"
    fi
else
    echo "iinit not found. Skipping iRODS authentication"
fi

echo "Running with arguments:" "$@"

exec "$@"
