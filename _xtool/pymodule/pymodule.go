package main


import (
	"log"
	"os"
	"fmt"
	"flag"
	"strings"
	_ "unsafe"
	"encoding/json"
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/py"
)

//go:linkname SequenceList C.PySequence_List
func SequenceList(o *py.Object) *py.Object { return nil }



type library struct {
	LibName 	string 		`json:"libName"`
	Depth 		int  		`json:"depth"`
	Modules 	[]string 	`json:"modules"`
}


// Python library name to module name mapping
var libToModule = map[string]string {
    "scikit-learn": "sklearn",
    "pillow":       "PIL",
}


func GetModuleName(libName string) string {
    if mod, ok := libToModule[libName]; ok {
        return mod
    }
    return libName
}

func (pkg *library) getModules(moduleName string, depth int) {
	if depth > pkg.Depth {
		return
	}
	mod := py.ImportModule(c.AllocaCStr(moduleName))
	if mod == nil {
		return
	}
	pkg.Modules = append(pkg.Modules, moduleName)
	if depth == pkg.Depth {
		return
	}
	pyPath := mod.GetAttrString(c.Str("__path__"))
	if pyPath == nil {
		return
	}
	pkgUtil := py.ImportModule(c.Str("pkgutil"))
	iterModules := pkgUtil.GetAttrString(c.Str("iter_modules"))
	iter := iterModules.Call(py.Tuple(pyPath), nil)
	subModules := SequenceList(iter)
	for i := 0; i < subModules.ListLen(); i++ {
		subModule := subModules.ListItem(i)
		name := subModule.TupleItem(1)
		nameStr := c.GoString(name.CStr())
		if strings.HasPrefix(nameStr, "test") || strings.HasPrefix(nameStr, "_") {
			continue
		}
		subModuleName := moduleName + "." + nameStr
		isPkg := subModule.TupleItem(2)
		if isPkg.IsTrue() == 1 {
			pkg.getModules(subModuleName, depth+1)
		} else {
			subMod := py.ImportModule(c.AllocaCStr(subModuleName))
			if subMod == nil {
				continue
			}
			pkg.Modules = append(pkg.Modules, subModuleName)
		}
	}
}

func main() {
	depth := flag.Int("d", 1, "extract depth")
	flag.Parse()
	if flag.NArg() < 1 {
        fmt.Fprintln(os.Stderr, "Usage: pymodule [-d <depth>] <libraryName>")
        os.Exit(1)
    }
	libraryName := flag.Arg(0)
	pkg := library{
		LibName: libraryName,
		Depth: *depth,
		Modules: []string{},
	}
	moduleName := GetModuleName(libraryName)
	pkg.getModules(moduleName, 1)
	if len(pkg.Modules) == 0 {
		log.Fatalf("error: import module failed: %s", moduleName)
	}
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		log.Fatalf("error: failed to marshal json: %v", err)
	}
	fmt.Println(string(data))
}
