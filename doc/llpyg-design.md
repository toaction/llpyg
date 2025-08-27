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

## 设计决策

### llpyg 是否需要脱离 LLGo ?
> https://github.com/goplus/llpyg/issues/5

llpyg 面向的是那些需要 LLGo Bindings 的用户，即 LLGo 开发者，因此可以依赖于 LLGo.

llpyg 依赖于 LLGo 的 Python 生态集成能力，该工具的一些子组件如 `pydump` 和 `pymodule` 需要 LLGo 进行编译和安装。

### llpyg 是否需要为用户提供与系统无关的 Python 环境？

> https://github.com/goplus/llpyg/issues/2#issuecomment-3200109475

llpyg 默认使用系统 Python 环境，并不为用户提供 Python 安装及第三方库自动下载的服务。

llpyg 只做 Python 环境的检查操作，当无法检测到系统的 Python 环境或检测到第三方库未安装时，触发 `panic`。

### 要转换的 Python 库的版本是否可以指定？

> https://github.com/goplus/llpyg/issues/8

llpyg 使用的是用户已经安装好的 Python 库的版本，并不支持对版本的修改。

用户若想转换不同版本的 Python 库，需要手动更改已安装的库。

### llpyg 是否需要为用户提供指定 Python 路径的功能？

> https://github.com/goplus/llpyg/issues/9

用户可以通过 `PYTHONHOME` 环境变量来指定 Python 路径。 llpyg 并不会提供一个单独的环境变量。


## 安装与使用

### Dependencies

- [LLGo](https://github.com/goplus/llgo)
- [Python 3.12+](https://www.python.org/)

### How to install
执行安装之前，请确保本地已存在 Python 环境。你可以通过 `PYTHONHOME` 来指定 Python 路径：
```bash
export PYTHONHOME=/path/to/python
```
若未设置 `PYTHONHOME`，你需要通过 `PKG_CONFIG_PATH` 环境变量来指定 Python 解释器的位置:
```bash
export PKG_CONFIG_PATH=/path/to/python/lib/pkgconfig:$PKG_CONFIG_PATH
```

**Install from source**:
```bash
git clone https://github.com/toaction/llpyg.git
cd llpyg
bash install.sh
```

### Usage

执行命令之前，请确保本地已安装要转换的 Python 第三方库，你可以通过 `PYTHONPATH` 环境变量来添加 Python 库的位置:
```bash
export PYTHONPATH=/path/to/lib:$PYTHONPATH
```
你可以选择两种不同的方式来执行命令，分别是：
- 命令行参数
- llpyg.cfg 配置文件

**1. 命令行参数**

```bash
llpyg [-o ouput_dir] [-mod mod_name] [-d module_depth] py_lib_name
```

- `-o`: LLGo Bindings output dir, default `./test`.
- `-mod`: Output Go module name, default `py_lib_name`.
- `-d`: Extract Python module max depth, default `1`.

**2. llpyg.cfg 文件**

```bash
llpyg [-o ouput_dir] [-mod mod_name] llpyg.cfg
```
llpyg.cfg 是配置文件，可以对内容进行修改，llpyg将会根据该文件执行程序。示例：

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

### Output

以 `numpy` 为例 (`-d=1`)，输出目录结构：

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



## How it works

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

- `_xtool`: 需要使用 LLGo 进行编译的子组件
- `cmd`: 可执行文件，每个子目录对应一个可执行文件
- `tool`: 仅用 Go 即可运行的子组件，应包含单元测试


### 函数签名解析
代码目录： `/tool/pysig`

从签名中提取四个信息：`Name`，`Type`, `DefaultValue`, `Optional`

支持可选项解析：
```Python
([start, ]stop, [step, ]dtype=None, *, device=None, like=None)
```
支持列表参数解析：
```Python
( (a1, a2, ...), axis=0, out=None, dtype=None, casting=\"same_kind\" )
```
支持复杂参数类型及默认值解析：
```Python
(start: 'Union[int, float]', stop: 'Union[int, float]', /, num: 'int', *, dtype: 'Optional[Dtype]' = None, device: 'Optional[Device]' = None, endpoint: 'bool' = True) -> 'Array'
```
解析结果：
```json
{
    "name": "start",
    "type": "'Union[int, float]'",
    "defVal": "",
    "optional": false
},
```

### 指定 Python 路径
当用户系统中存在 `PYTHONHOME` 环境变量时，llpyg 会优先使用该环境变量指定的 Python 路径。

在安装 llpyg 时，需要用到 LLGo 的 Python 集成能力，此时需要为 LLGo 指定 Python 解释器的位置。因此在安装时，需要设置一些临时的系统环境变量：
```bash
# install.sh

pyHome=$PYTHONHOME
if [ -n "$pyHome" ]; then
    export PKG_CONFIG_PATH="$pyHome/lib/pkgconfig:$PKG_CONFIG_PATH"
fi
```
现阶段 Python 库的位置是动态指定的，在每次运行时，都需要重新设置环境变量。
```bash
# /tool/pyenv/prepare.go

export PATH="$pyHome/bin:$PATH"
export PKG_CONFIG_PATH="$pyHome/lib/pkgconfig:$PKG_CONFIG_PATH"
export DYLD_LIBRARY_PATH="$pyHome/lib:$DYLD_LIBRARY_PATH"
```

### 符号信息提取
代码目录： `/_xtool/pydump`
```bash
pydump py_module_name
```
从 Python 模块中提取符号信息，包括：
- Name
- Type
- Doc
- Signature

对于函数，签名的提取方法：
1. 利用反射 `inspect.signature` 获取 Python 函数签名
2. 从文档中获取函数签名，该方法适用于 C 编写的内置函数
3. 将签名统一设置为 `(*args, **kwargs)`


