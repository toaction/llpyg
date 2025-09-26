package demo

import (
	"github.com/goplus/lib/py"
	_ "unsafe"
)

const LLGoPackage = "py.demo"

//go:linkname FuncA py.func_a
func FuncA() *py.Object

//go:linkname FuncB py.func_b
func FuncB() *py.Object

//go:linkname FuncC py.func_c
func FuncC(a *py.Object, b *py.Object) *py.Object

//go:linkname FuncD py.func_d
func FuncD(a *py.Object, b *py.Object) *py.Object

//go:linkname FuncE py.func_e
func FuncE(start *py.Object, unit *py.Object) *py.Object
