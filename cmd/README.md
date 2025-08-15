requirements:
```text
llgo
Python 3.12
```
## Run
install tools
```bash
cd _xtool
llgo install ./...      # pydump, pymodule

cd llpyg
go install -v ./cmd/...     # llpyg
```
run `llpyg`
```bash
llpyg python_lib[==version]
```
args: 
 - `-o`: Output dir, default `./test`
 - `-mod`: LLGo Bindings module name, default `libName`
 - `-d`: Extract Python module depth, default `1`

example:
```bash
llpyg -o ./python_lib -mod github.com/llpyg/numpy numpy==2.3.0
```
Generated files:
```go
./python_lib/numpy
├── go.mod
├── go.sum
├── llpyg.cfg     // llpyg config file  
└── numpy.go      // LLGo Bindings for numpy 
```