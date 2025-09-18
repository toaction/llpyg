package var_

import (
	"github.com/goplus/lib/py"
	_ "unsafe"
)

const LLGoPackage = "py.var"

//go:linkname A py.A
var A *py.Object

//go:linkname A_ py.a
var A_ *py.Object

//go:linkname B py.b
var B *py.Object

//go:linkname C py.c
var C *py.Object

//go:linkname D py.d
var D *py.Object

//go:linkname S py.s
var S *py.Object

//go:linkname L py.l
var L *py.Object

//go:linkname T py.t
var T *py.Object

//go:linkname N py.n
var N *py.Object
