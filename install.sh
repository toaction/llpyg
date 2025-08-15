export PYTHONHOME=$LLGO_ROOT/python
export PATH=$PYTHONHOME/bin:$PATH
export DYLD_LIBRARY_PATH=$PYTHONHOME/lib
export PKG_CONFIG_PATH=$PYTHONHOME/lib/pkgconfig

cd ./_xtool
llgo install ./...

cd ../
go install -v ./cmd/...