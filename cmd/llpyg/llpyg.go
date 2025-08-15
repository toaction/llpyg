package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"go/token"
	"log"
	"strconv"
	"go/ast"
	"go/types"
	"os/exec"
	"strings"
	"encoding/json"
	"path/filepath"
	"github.com/goplus/gogen"
	"github.com/goplus/llpyg/tool/pyenv"
	"github.com/goplus/llpyg/tool/pysig"
	_ "github.com/goplus/lib/py"
)


type Config struct {
	Name 		string 		`json:"name"`			// go module name
	LibName 	string 		`json:"libName"`		// Python library name
	LibVersion 	string 		`json:"libVersion"`		// Python library version
	Modules		[]string 	`json:"modules"`		// Python modules
}


type symbol struct {
	Name string `json:"name"`		// python name
	Type string `json:"type"`
	Doc  string `json:"doc"`
	Sig  string `json:"sig"`
}

type module struct {
	Name  string    `json:"name"`		// python module name
	Items []*symbol `json:"items"`
}

type library struct {
	LibName 	string 		`json:"libName"`
	Depth 		int  		`json:"depth"`
	Modules 	[]string 	`json:"modules"`
}

func pydump(moduleName string) (mod module, err error) {
    var out bytes.Buffer
    cmd := exec.Command("pydump", moduleName)
    cmd.Stdout = &out
    cmd.Stderr = os.Stderr
    err = cmd.Run()
    if err != nil {
        return mod, fmt.Errorf("pydump %s failed", moduleName)
    }
    err = json.Unmarshal(out.Bytes(), &mod)
    if err != nil {
        return mod, fmt.Errorf("unmarshal %s failed", moduleName)
    }
	if mod.Name != moduleName {
		return mod, fmt.Errorf("import module failed: %s", moduleName)
	}
    return mod, nil
}

func pymodule(libName string, depth int) (lib library, err error) {
	var out bytes.Buffer
	cmd := exec.Command("pymodule", "-d", strconv.Itoa(depth), libName)
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return lib, fmt.Errorf("pymodule %s failed", libName)
	}
	err = json.Unmarshal(out.Bytes(), &lib)
	if err != nil {
		return lib, fmt.Errorf("unmarshal %s failed", libName)
	}
	if len(lib.Modules) == 0 {
		return lib, fmt.Errorf("get modules from package %s failed", libName)
	}
	return lib, nil
}



func main() {
	var libName, libVersion string

	output := flag.String("o", "./test", "Output dir")
	modName := flag.String("mod", "", "Generate Go Bindings module name")
	modDepth := flag.Int("d", 1, "Extract module depth")
	flag.Parse()

    if flag.NArg() < 1 {
        fmt.Fprintln(os.Stderr, "Usage: llpyg [-o outputDir] [-mod modName] [-d modDepth] pythonLibName[==version]")
        os.Exit(1)
    }
	libArg := flag.Arg(0)
	libName, libVersion = getNameAndVersion(libArg)
	if libName == "" {
		fmt.Fprintln(os.Stderr, "error: Python library name cannot be empty")
		os.Exit(1)
	}
	fmt.Printf("llpyg args: libName=%s, libVersion=%s, output=%s, modName=%s, modDepth=%d\n", libName, libVersion, *output, *modName, *modDepth)

	// check Python environment and library
	fmt.Printf("Checking Python environment...\n")
	envChecked, version := pyenv.PyEnvCheck(libName, libVersion)
	if !envChecked {
		os.Exit(1)
	}
	fmt.Printf("Python library (%s %s) is ready\n", libName, version)

	// generate llpyg.cfg
	lib, err := pymodule(libName, *modDepth)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cfg := Config{
		Name: 		lib.Modules[0],		// go package name
		LibName: 	libName,
		LibVersion: version,
		Modules: 	lib.Modules,
	}
	outDir := filepath.Join(*output, cfg.Name)
	err = os.RemoveAll(outDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := writeConfig(cfg, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	
	// go module
	if *modName == "" {
		*modName = cfg.Name
	}
	genGoModule(*modName, outDir)

	// LLGo Bindings generation
	generateFromConfig(cfg, outDir)

	fmt.Printf("LLGo bindings generated successfully in %s\n", outDir)
}

func generateFromConfig(cfg Config, outDir string) {
	for _, moduleName := range cfg.Modules {
		fmt.Printf("Generating Go bindings for %s...\n", moduleName)
		genGoBindings(moduleName, outDir)
	}
}


func genGoBindings(moduleName, outDir string) {
	// pydump
	mod, err := pydump(moduleName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// create go package
	parts := strings.Split(mod.Name, ".")
	pkgName := parts[len(parts)-1]
	pkg := gogen.NewPackage("", pkgName, nil)
	pkg.Import("unsafe").MarkForceUsed(pkg)      // import _ "unsafe"
	py := pkg.Import("github.com/goplus/lib/py") // import "github.com/goplus/lib/py"
	
	f := func(cb *gogen.CodeBuilder) int {
		cb.Val("py." + mod.Name)
		return 1
	}
	defs := pkg.NewConstDefs(pkg.Types.Scope())
	defs.New(f, 0, 0, nil, "LLGoPackage")		// const LLGoPackage = "py.moduleName"

	obj := py.Ref("Object").(*types.TypeName).Type().(*types.Named)
	objPtr := types.NewPointer(obj)
	ret := types.NewTuple(pkg.NewParam(0, "", objPtr))   // return *py.Object
	
	// add context
	ctx := &context{pkg, obj, objPtr, ret, nil, nil, py}
	ctx.genMod(pkg, &mod)
	skips := ctx.skips
	if n := len(skips); n > 0 {
		log.Printf("==> Skip %d symbols:\n%v\n", n, skips)
	}

	// write to file
    outputFile := moduleToPath(mod.Name)
    file, err := createFileWithDirs(outputFile)
    if err != nil {
        log.Printf("error: failed to create output file %s: %v\n", outputFile, err)
        return
    }
    defer file.Close()

    err = pkg.WriteTo(file)
    if err != nil {
        log.Printf("error: failed to write Go code to file %s: %v\n", outputFile, err)
        return
    }
}


func moduleToPath(moduleName string) string {
    parts := strings.Split(moduleName, ".")
	fileName := parts[len(parts)-1] + ".go"
    parts = parts[1:]
	path := strings.Join(parts, "/")
    return filepath.Join(path, fileName)
}


type context struct {
	pkg    *gogen.Package
	obj    *types.Named
	objPtr *types.Pointer
	ret    *types.Tuple
	skips  []element
	todos  []element
	py     gogen.PkgRef
}

type element struct {
	name 	string
	pyType 	string
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

func (ctx *context) genMod(pkg *gogen.Package, mod *module) {
	// 重复函数覆盖
	funcMap := make(map[string]symbol)
	for _, sym := range mod.Items {
		if inFuncSet(sym.Type) {
			funcMap[sym.Name] = *sym
			continue
		}
		ctx.todos = append(ctx.todos, element{name: sym.Name, pyType: sym.Type})
	}
	for _, sym := range funcMap {
		ctx.genFunc(pkg, &sym)
	}
}

func (ctx *context) genFunc(pkg *gogen.Package, sym *symbol) {
	name, symSig := sym.Name, sym.Sig
	if len(name) == 0 || name[0] == '_' {
		return
	}
	if symSig == "" {		// no signature
		ctx.skips = append(ctx.skips, element{name: name, pyType: sym.Type})
		return
	}
	// signature
	params, variadic := ctx.genParams(pkg, symSig)
	name = genName(name, -1)
	sig := types.NewSignatureType(nil, nil, nil, params, ctx.ret, variadic)		// ret: *py.Object
	fn := pkg.NewFuncDecl(token.NoPos, name, sig)
	// doc
	docList := ctx.genDoc(sym.Doc)
	if len(docList) > 0 {
		docList = append(docList, emptyCommentLine)
	}
	docList = append(docList, ctx.genLinkname(name, sym))
	fn.SetComments(pkg, &ast.CommentGroup{List: docList})
}

func (ctx *context) genParams(pkg *gogen.Package, sig string) (*types.Tuple, bool) {
	args := pysig.Parse(sig)
	if len(args) == 0 {
		return nil, false
	}
	n := len(args)
	objPtr := ctx.objPtr
	list := make([]*types.Var, 0, n)
	for i := 0; i < n; i++ {
		name := args[i].Name
		if name == "/" {
			continue
		}
		if name == "*" || name == "\\*" {
			break
		}
		if strings.HasPrefix(name, "*") {
			if name[1] != '*' {
				list = append(list, vArgs)
				return types.NewTuple(list...), true
			}
			return types.NewTuple(list...), false
		}
		list = append(list, pkg.NewParam(0, genName(name, 0), objPtr))
	}
	return types.NewTuple(list...), false
}


// round round_
func genName(name string, idxDontTitle int) string {

    lastIdx := len(name) - 1
    for lastIdx >= 0 && name[lastIdx] == '_' {
        lastIdx--
    }
	workingName := name[:lastIdx+1]
	trail := name[lastIdx+1:]
    parts := strings.Split(workingName, "_")

	for i, part := range parts {
		if i != idxDontTitle && part != "" {
			if c := part[0]; c >= 'a' && c <= 'z' {
				part = string(c+'A'-'a') + part[1:]
			}
			parts[i] = part
		}
	}
	name = strings.Join(parts, "")
	switch name {
	case "default", "func", "var", "range", "":
		name += "_"
	}
	return name + trail
}

func (ctx *context) genLinkname(name string, sym *symbol) *ast.Comment {
	return &ast.Comment{Text: "//go:linkname " + name + " py." + sym.Name}
}


// Generate documentation comments from the symbol's doc string
func (ctx *context) genDoc(doc string) []*ast.Comment {
	if doc == "" {
		return make([]*ast.Comment, 0, 4)
	}
	lines := strings.Split(doc, "\n")
	list := make([]*ast.Comment, len(lines), len(lines)+4)
	for i, line := range lines {
		list[i] = &ast.Comment{Text: "// " + line}
	}
	return list
}



const (
	NameValist = "__llgo_va_list"
)

var (
	emptyCommentLine = &ast.Comment{Text: "//"}
	tyAny = types.NewInterfaceType(nil, nil)
	vArgs = types.NewParam(0, nil, NameValist, types.NewSlice(tyAny))
)

