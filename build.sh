#!/bin/bash

# Load environment variables from .env file
export $(grep -v '^#' .env | xargs)

# Build the Go binary with the environment variables
go build -ldflags "\
    -X 'main.omdbAPIKey=${OMDB_API_KEY}' \
    -X 'main.openSubtitlesAPIKey=${OPENSUBTITLES_API_KEY}' \
    -X 'main.openSubtitlesUsername=${OPENSUBTITLES_USERNAME}' \
    -X 'main.openSubtitlesPassword=${OPENSUBTITLES_PASSWORD}'"