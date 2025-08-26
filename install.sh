#!/bin/bash

if [ -z "$LLGO_ROOT" ]; then
    echo "LLGO_ROOT is not set"
    exit 1
fi

if [ -n "$PYTHONHOME" ]; then
    export PKG_CONFIG_PATH="$PYTHONHOME/lib/pkgconfig:$PKG_CONFIG_PATH"
fi

echo "Installing llpyg..."

cd ./_xtool
llgo install ./...

cd ../
go install -v ./cmd/...
