## 基本介绍

**[LLGo](https://github.com/goplus/llgo)**：一个基于LLVM的 Go 编译器，以便更好地与 C 和 Python 生态集成。

**LLGo Bindings**：为了能够在 Go 代码中调用其他语言，LLGo 需要先将其他语言的符号进行映射，映射为 Go 符号，我们将这种一一映射的符号声明称之为 LLGo Bindings. 例如 Python 的 `numpy.add` 函数:
```Python
numpy.add(x1, x2, /, out=None, *, where=True, casting='same_kind', order='K', dtype=None, subok=True[, signature, extobj])
```

对应的 LLGo Bindings 为：
```Go
//go:linkname Add py.add
func Add(x1 *py.Object, x2 *py.Object) *py.Object
```

**[llpyg](https://github.com/goplus/llpyg)**：一个面向 Python 库的 LLGo Bindings 自动生成工具。

## 前提条件

### llpyg 面向人群

- LLGo 开发者

### 安装前提

- Go
- [LLGo](https://github.com/goplus/llgo)

llpyg 需要 LLGo 对一些子组件进行编译，如 `pydump` 和 `pymodule`. 但当 llpyg 安装完成后，无需 LLGo 即可运行。

### 运行前提

- Python 3.12 (通过 `python3` 或 `python` 检测)
- 要转换的第三方库 (通过 `pip3 show` 或 `pip show` 检测)

## 安装与使用

### Install

```bash
bash install.sh
```

### Usage

**1. 命令行**

```bash
llpyg [-o ouput_dir] [-mod mod_name] [-d module_depth] py_lib_name
```

- `-o`: LLGo Bindings output dir, default `./test`.
- `-mod`: Output Go module name, default `py_lib_name`.
- `-d`: Extract Python module max depth, default `1`.

**2. llpyg.cfg 文件**

```json
{
  "name": "numpy",
  "libName": "numpy",
  "modules": [
    "numpy"
  ]
}
```

- `name`: Go package name.
- `libName`: Python library name.
- `modules`: Extract Python modules.

```bash
llpyg [-o ouput_dir] [-mod mod_name] llpyg.cfg
```

### Output

输出目录结构：以 `numpy` 为例

```go
numpy
├── go.mod
├── go.sum
├── llpyg.cfg				// config 配置文件
└── numpy.go				// LLGo Bindings
```

`numpy.go` 文件 (LLGo Bindings)：

```go
package numpy

import (
	"github.com/goplus/lib/py"
	_ "unsafe"
)

const LLGoPackage = "py.numpy"

//go:linkname Maximum py.maximum
func Maximum(x1 *py.Object, x2 *py.Object) *py.Object
```


## 开发相关

### 项目结构

```go
llpyg/
├── _xtool/
│   ├── pkg1/
│   └── pkg2/
├── cmd/
│   └── llpyg/
│       └── llpyg.go		
├── tool/
│   ├── pkg1/
│   └── pkg2/
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

- `_xtool`: 需要使用 LLGo 进行编译的子命令
- `cmd`: 可执行文件，每个子目录对应一个可执行文件
- `tool`: 子组件，应对每一个可独立的子组件进行单元测试





























