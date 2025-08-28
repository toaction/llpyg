#!/bin/bash

# exit on error
set -e

echo "Installing llpyg..."

cd ./_xtool
llgo install ./...

cd ../
go install -v ./cmd/...

echo "llpyg is now available in your GOPATH."
