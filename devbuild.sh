#!/usr/bin/env bash

HOME=/Users/richardmauri
APP_DIR="$HOME/projects/homepage"
APP="$APP_DIR/dist/homepage-darwin-amd64"

go test ./...
go build -o $APP .
