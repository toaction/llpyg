package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/py"
	"github.com/goplus/lib/py/inspect"
)



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


// write module instance into JSON file
func dumpToJson(mod *module, outDir string) error {
	// Convert the module to JSON and write it to a file
	filePath := filepath.Join(outDir, mod.Name + ".json")
	file, err := createFileWithDirs(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mod); err != nil {
		return err
	}
	return nil
}