package main

import (
	"os"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/py"
	"github.com/goplus/lib/py/inspect"
)


type symbol struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
	Sig  string `json:"sig"`
}

type module struct {
	Name  string    `json:"name"`		// python module name
	Items []*symbol `json:"items"`
}


var funcSet = []string {
	"ufunc",
	"method",
	"function", 
	"method-wrapper",
	"builtin_function_or_method",
	"_ArrayFunctionDispatcher",
}


func inFuncSet(typeName string) bool {
	for _, item := range funcSet {
		if item == typeName {
			return true
		}
	}
	return false
}

func extractSignatureFromDoc(doc, funcName string) string {
    lines := strings.SplitN(doc, "\n\n", 2)
    if len(lines) == 0 {
        return ""
    }
    firstLine := strings.TrimSpace(lines[0])
    if !strings.HasPrefix(firstLine, funcName+"(") {
        return ""
    }
	idx := strings.Index(firstLine, "(")
    if idx == -1 {
        return ""
    }
    params := firstLine[idx:]
    fields := strings.Fields(params)
    return strings.Join(fields, " ")
}


func getSignature(val *py.Object, sym *symbol) string {
	// function, method, class, or implement __call__
	if val.Callable() == 0 {
		return ""
	}
	// get signature from inspect
	sigFromInspect := inspect.Signature(val)
	if sigFromInspect != nil {
		sig := c.GoString(sigFromInspect.Str().CStr())
		if sig != "(*args, **kwargs)" {
			return sig
		}
	}
	// get signature from doc
	sigFromDoc := extractSignatureFromDoc(sym.Doc, sym.Name)
	if sigFromDoc != "" {
		return sigFromDoc
	}
	// Paradigms
	if inFuncSet(sym.Type) {
		return "(*args, **kwargs)"
	}
	return ""
}


// moduleName: Python module name
func pydump(moduleName string) (*module, error) {
	// import module
	// mod := py.ImportModule(c.Str(moduleName)) // panic: cstr(<string-literal>): invalid arguments
	mod := py.ImportModule(c.AllocaCStr(moduleName))
	if mod == nil {
		return nil, fmt.Errorf("failed to import module: %s", moduleName)
	}
	// get dict, python list Object
	keys := mod.ModuleGetDict().DictKeys()
	if keys == nil {
		return nil, fmt.Errorf("failed to get module dict keys: %s", moduleName)
	}
	// create module instance
	modInstance := &module{
		Name: moduleName,
	}
	// get symbols
	for i, n := 0, keys.ListLen(); i < n; i++ {
		key := keys.ListItem(i)
		val := mod.GetAttr(key)
		if val == nil {
			continue
		}
		// name, type, doc, signature
		sym := &symbol{}
		sym.Name = c.GoString(key.CStr())
		sym.Type = c.GoString(val.Type().TypeName().CStr())
		// doc
		doc := val.GetAttrString(c.Str("__doc__"))
		if doc != nil {
			// sym.Doc = c.GoString(doc.CStr())	// panic: signal: segmentation fault
			sym.Doc = c.GoString(doc.Str().CStr())
		}
		// signature
		sym.Sig = getSignature(val, sym)
		modInstance.Items = append(modInstance.Items, sym)
	}
	return modInstance, nil
}


func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pydump <py_module_name>")
		return
	}
	// get module name from command line argument
	moduleName := os.Args[1]
	mod, err := pydump(moduleName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// print module information
	data, err := json.MarshalIndent(mod, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}