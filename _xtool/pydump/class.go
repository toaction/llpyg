package main

import (
	"fmt"
	"strings"

	"github.com/goplus/lib/c"
	"github.com/goplus/lib/py"
	"github.com/goplus/lib/py/inspect"
	"github.com/goplus/llpyg/symbol"
)

// get builtins.dict function object
// for getting real dict of mappingproxy
func getBuiltinDict() (*py.Object, error) {
	builtins := py.ImportModule(c.AllocaCStr("builtins"))
	if builtins == nil {
		return nil, fmt.Errorf("can't import builtins")
	}
	dictFunc := builtins.GetAttrString(c.Str("dict"))
	if dictFunc == nil {
		return nil, fmt.Errorf("can't get dict from builtins")
	}
	return dictFunc, nil
}


func parseClass(pycls *py.Object, sym *symbol.Symbol, modName string, dictFunc *py.Object) (*symbol.Class, error) {
	cls := &symbol.Class{
		Name: sym.Name,
		Doc: sym.Doc,
	}
	clsModule := c.GoString(pycls.GetAttrString(c.Str("__module__")).CStr())
	if clsModule != modName {
		return nil, fmt.Errorf("%s is not in module %s", sym.Name, modName)
	}
	// bases
	bases := pycls.GetAttrString(c.Str("__bases__"))   // tuple
	if bases == nil {
		return nil, fmt.Errorf("can't get __bases__ of %s", sym.Name)
	}
	for i, n := 0, bases.TupleLen(); i < n; i++ {
		baseObj := bases.TupleItem(i)
		base := &symbol.Base{
			Name: c.GoString(baseObj.GetAttrString(c.Str("__name__")).CStr()),
			Module: c.GoString(baseObj.GetAttrString(c.Str("__module__")).CStr()),
		}
		cls.Bases = append(cls.Bases, base)
	}
	// methods and properties
	cls, err := parseClassDict(pycls, cls, dictFunc)
	if err != nil {
		return nil, err
	}
	return cls, nil
}

func parseClassDict(pycls *py.Object, cls *symbol.Class, dictFunc *py.Object) (*symbol.Class, error) {
	dict := pycls.GetAttrString(c.Str("__dict__"))
	if dict == nil {
		return nil, fmt.Errorf("can't get __dict__ of %s", cls.Name)
	}
	dictTypeName := c.GoString(dict.Type().TypeName().CStr())
	if dictTypeName != "mappingproxy" {
		return nil, fmt.Errorf("__dict__ of %s is not a mappingproxy", cls.Name)
	}
	realDict := dictFunc.CallOneArg(dict)
	items := realDict.DictItems()
	for i, n := 0, items.ListLen(); i < n; i++ {
		item := items.ListItem(i)
		name := c.GoString(item.TupleItem(0).CStr())
		val := item.TupleItem(1)
		typeName := c.GoString(val.Type().TypeName().CStr())
		typeName = strings.TrimSpace(typeName)
		if name == "__init__" || name == "__new__" {
			continue
		}
		switch typeName {
		case "function":
			sym, err := parseMethod(val, name, typeName)
			if err != nil {
				return nil, err
			}
			cls.InstanceMethods = append(cls.InstanceMethods, sym)
		case "classmethod", "staticmethod":
			val = val.GetAttrString(c.Str("__func__"))
			sym, err := parseMethod(val, name, typeName)
			if err != nil {
				return nil, err
			}
			if typeName == "classmethod" {
				cls.ClassMethods = append(cls.ClassMethods, sym)
			} else {
				cls.StaticMethods = append(cls.StaticMethods, sym)
			}
		case "property":
			property := parseProperty(val, name)
			cls.Properties = append(cls.Properties, property)
		default:
			// TODO: others
		}
		// init method
	}
	return cls, nil
}

func parseMethod(val *py.Object, name string, typeName string) (*symbol.Symbol, error) {
	sym := &symbol.Symbol{
		Name: name,
		Type: typeName,
	}
	doc := val.GetAttrString(c.Str("__doc__"))
	if doc != nil && doc.IsTrue() == 1 {
		sym.Doc = c.GoString(doc.CStr())
	}
	sig, err := getSignature(val, sym)
	if err != nil {
		return nil, fmt.Errorf("can't get signature of %s", name)
	}
	sym.Sig = *sig
	return sym, nil
}

func parseProperty(val *py.Object, name string) *symbol.Property {
	property := &symbol.Property{
		Name: name,
	}
	getter := val.GetAttrString(c.Str("fget"))
	if getter != nil && getter.IsTrue() == 1 {
		property.Getter = symbol.Signature{}
	}
	setter := val.GetAttrString(c.Str("fset"))
	if setter != nil && setter.IsTrue() == 1 {
		sig := inspect.Signature(setter)
		if sig != nil {
			property.Setter = symbol.Signature{
				Source: symbol.SigSourceInspect,
				Str:    c.GoString(sig.Str().CStr()),
			}
		}
	}
	return property
}
