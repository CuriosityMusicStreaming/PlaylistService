#!/bin/bash

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
PROJECT_DIR=$(dirname "$SCRIPT_DIR")

pushd "$PROJECT_DIR" || exit
docker-compose --project-directory "$PROJECT_DIR" -f data/docker/tests/docker-compose.yml up --build --abort-on-container-exit --exit-code-from playlistservice-api-client
popd || exit
