# !/bin/bash

pyHome=$PYTHONHOME
if [ -z "$pyHome" ]; then
    pyHome=$LLPYG_PYHOME
fi

echo "Python Home: $pyHome"

if [ -n "$pyHome" ]; then
    export PKG_CONFIG_PATH="$pyHome/lib/pkgconfig:$PKG_CONFIG_PATH"
fi

echo "Install llpyg..."

cd ./_xtool
llgo install ./...
cd ../
go install -v ./cmd/...
