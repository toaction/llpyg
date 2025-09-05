package main

import (
	"os"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/goplus/lib/py"
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/py/inspect"
)


type symbol struct {
	Name 		string	`json:"name"`
	TypeName 	string	`json:"type"`
	Doc 		string	`json:"doc"`
	Sig 		string	`json:"sig"`
}

type baseClass struct {
	Name 	string	`json:"name"`
	Module 	string	`json:"module"`
}

type class struct {
	Name 			string			`json:"name"`
	Module 			string			`json:"module"`
	Bases 			[]*baseClass	`json:"bases"`
	Attributes  	[]string		`json:"attributes"`
	Properties  	[]string		`json:"properties"`
	InitMethod 		*symbol			`json:"initMethod"`
	InstanceMethods []*symbol		`json:"instanceMethods"`
	ClassMethods 	[]*symbol		`json:"classMethods"`
	StaticMethods 	[]*symbol		`json:"staticMethods"`
	// specialMethods 	[]*symbol		// into instanceMethods
}

type moduleInfo struct {
	Name 	string		`json:"name"`
	Classes []*class	`json:"classes"`
}


func parseMethod(val *py.Object, name string, typeName string) *symbol {
	sym := &symbol{Name: name, TypeName: typeName}
	sigObj := inspect.Signature(val)
	if sigObj != nil {
		sym.Sig = c.GoString(sigObj.Str().CStr())
	}
	docObj := inspect.Getdoc(val)
	if docObj != nil {
		sym.Doc = c.GoString(docObj.Str().CStr())
	}
	return sym
}


func parseClass(clsObj *py.Object) *class {
	cls := &class{}
	cls.Name = c.GoString(clsObj.GetAttrString(c.Str("__name__")).CStr())
	cls.Module = c.GoString(clsObj.GetAttrString(c.Str("__module__")).CStr())
	basesObj := clsObj.GetAttrString(c.Str("__bases__"))   // tuple
	cls.Bases = make([]*baseClass, 0)
	for i, n := 0, basesObj.TupleLen(); i < n; i++ {
		baseObj := basesObj.TupleItem(i)
		base := &baseClass{
			Name: c.GoString(baseObj.GetAttrString(c.Str("__name__")).CStr()),
			Module: c.GoString(baseObj.GetAttrString(c.Str("__module__")).CStr()),
		}
		cls.Bases = append(cls.Bases, base)
	}
	dict := clsObj.GetAttrString(c.Str("__dict__"))
	dictTypeName := c.GoString(dict.Type().TypeName().CStr())
	if dictTypeName != "mappingproxy" {
		return nil
	}
	builtins := py.ImportModule(c.AllocaCStr("builtins"))
	dictFunc := builtins.GetAttrString(c.Str("dict"))
	realDict := dictFunc.CallOneArg(dict)
	items := realDict.DictItems()
	for i, n := 0, items.ListLen(); i < n; i++ {
		item := items.ListItem(i)
		name := c.GoString(item.TupleItem(0).CStr())
		val := item.TupleItem(1)
		typeName := c.GoString(val.Type().TypeName().CStr())
		typeName = strings.TrimSpace(typeName)
		if name == "__init__" {
			sym := parseMethod(val, name, typeName)
			cls.InitMethod = sym
			continue
		}
		switch typeName {
		case "function":
			sym := parseMethod(val, name, typeName)
			cls.InstanceMethods = append(cls.InstanceMethods, sym)
		case "classmethod":
			val = val.GetAttrString(c.Str("__func__"))
			sym := parseMethod(val, name, typeName)
			cls.ClassMethods = append(cls.ClassMethods, sym)
		case "staticmethod":
			val = val.GetAttrString(c.Str("__func__"))
			sym := parseMethod(val, name, typeName)
			cls.StaticMethods = append(cls.StaticMethods, sym)
		case "property":
			cls.Properties = append(cls.Properties, name)
		default:
			cls.Attributes = append(cls.Attributes, name)
		}
	}
	return cls
}


func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pyclass <py_module_name>")
		return
	}
	mod := &moduleInfo{
		Name: os.Args[1],
	}
	modObj := py.ImportModule(c.AllocaCStr(mod.Name))
	if modObj == nil {
		fmt.Fprintf(os.Stderr, "Module %s not found\n", mod.Name)
		os.Exit(1)
	}
	keys := modObj.ModuleGetDict().DictKeys()
	for i, n := 0, keys.ListLen(); i < n; i++ {
		key := keys.ListItem(i)
		val := modObj.GetAttr(key)
		if val == nil {
			continue
		}
		typeName := c.GoString(val.Type().TypeName().CStr())
		if typeName == "type" {
			cls := parseClass(val)
			if cls != nil {
				mod.Classes = append(mod.Classes, cls)
			}
		}
	}
	data, err := json.MarshalIndent(mod, "", "  ")
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to marshal json: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(data))
}
