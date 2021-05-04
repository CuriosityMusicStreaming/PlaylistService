#!/usr/bin/env bash

set -o errexit

PARENT_DIR=$(dirname "$(readlink -f "$0")")
PROJECT_DIR=$(dirname "$PARENT_DIR")
OUTPUT_DIR=${PROJECT_DIR}/data/mysql/migrations

if [ -z "$1" ]; then
    echo "error: missing migration name" 1>&2
    echo "Usage:" 1>&2
    echo "  $(basename "$0") <migration-name>" 1>&2
    echo "Example:" 1>&2
    echo "  $(basename "$0") create_user" 1>&2
    exit 1
fi

# Datetime without delimiters, e.g. 20190531075026 (same as 2019-05-31 07:50:26)
DATETIME_PREFIX=$(date -u "+%Y%m%d%H%M%S")
MIGRATION_NAME=$(printf "%s" "$1" | tr -c "a-z0-9" "_")

writeSectionIntoFile() {
  echo "-- +migrate $1" >> "$2"
}

touchMigrationFile() {
    local FILENAME=${DATETIME_PREFIX}-${MIGRATION_NAME}.sql
    local FILEPATH=${OUTPUT_DIR}/${FILENAME}
    touch "$FILEPATH"
    writeSectionIntoFile "Up" "$FILEPATH"
    echo "" >> "$FILEPATH"
    writeSectionIntoFile "Down" "$FILEPATH"
    echo "created $FILEPATH"
}

touchMigrationFile
echo "now you can write SQL migrations in these file"
