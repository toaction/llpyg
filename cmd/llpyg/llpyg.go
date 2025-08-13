package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"go/token"
	"log"
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
	Name 		string `json:"name"`			// go module name
	LibName 	string `json:"libName"`			// Python library name
	LibVersion 	string `json:"libVersion"`		// Python library version
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


// numpy==2.3.0  ===> numpy, 2.3.0
func getNameAndVersion(arg string) (string, string) {
	parts := strings.SplitN(arg, "==", 2)
    name := parts[0]
    version := ""
    if len(parts) == 2 {
        version = parts[1]
    }
	return name, version
}

func createFileWithDirs(filePath string) (*os.File, error) {
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, err
    }
    return os.Create(filePath)
}


func writeConfig(cfg Config, outDir string) error {
	cfgPath := filepath.Join(outDir, "llpyg.cfg")
    file, err := createFileWithDirs(cfgPath)
    if err != nil {
        return fmt.Errorf("error: failed to create file: %w", err)
    }
    defer file.Close()
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(cfg); err != nil {
        return fmt.Errorf("error: failed to write configuration: %w", err)
    }
	return nil
}

func main() {
	// python library name and version, main module name
	// e.g. scikit-learn==1.0.2 : scikit-learn, 1.0.2, sklearn
	var libName, libVersion, moduleName string

	output := flag.String("o", "./test", "Output dir")
	modName := flag.String("mod", "", "Generate Go Bindings module name")
	flag.Parse()

	fmt.Printf("llpyg args: output=%s, modName=%s\n", *output, *modName)

	// get library name and version from command line argument
    if flag.NArg() < 1 {
        fmt.Fprintln(os.Stderr, "Usage: lpyg [-o outputDir] [-mod modName] pythonLibName[==version]")
        os.Exit(1)
    }
	libArg := flag.Arg(0)
	libName, libVersion = getNameAndVersion(libArg)
	if libName == "" {
		fmt.Fprintln(os.Stderr, "error: Python library name cannot be empty")
		os.Exit(1)
	}

	// check Python environment and library
	fmt.Printf("Checking Python environment...\n")
	envChecked, version := pyenv.PyEnvCheck(libName, libVersion)
	if !envChecked {
		os.Exit(1)
	}
	fmt.Printf("Python library (%s %s) is ready\n", libName, version)

	// generate llpyg.cfg
	moduleName = pyenv.GetModuleName(libName)
	cfg := Config{
		Name: 		strings.ReplaceAll(moduleName, "-", "_"),		// go package name
		LibName: 	libName,
		LibVersion: version,
	}
	outDir := filepath.Join(*output, cfg.Name)
	if err := writeConfig(cfg, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// LLGo Bindings generation
	generateFromConfig(cfg, outDir)

	// go module
	if *modName != "" {
		genGoModule(*modName, outDir)
	}

	fmt.Printf("LLGo bindings generated successfully in %s\n", outDir)
}


func genGoModule(modName string, outDir string) {
    if err := os.Chdir(outDir); err != nil {
		log.Printf("error: failed to change directory to %s: %v\n", outDir, err)
        os.Exit(1)
    }
    // init go module
    cmd := exec.Command("go", "mod", "init", modName)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        log.Printf("error: failed to initialize Go module: %v\n", err)
        os.Exit(1)
    }
    // go mod tidy
    tidyCmd := exec.Command("go", "mod", "tidy")
    tidyCmd.Stdout = os.Stdout
    tidyCmd.Stderr = os.Stderr
    if err := tidyCmd.Run(); err != nil {
        log.Printf("error: failed to run go mod tidy: %v\n", err)
		os.Exit(1)
    }
}


func pydump(pyModuleName string) (mod module, err error) {
    var out bytes.Buffer
    cmd := exec.Command("pydump", pyModuleName)
    cmd.Stdout = &out
    cmd.Stderr = os.Stderr
    err = cmd.Run()
    if err != nil {
        return mod, fmt.Errorf("failed to execute pydump %s command: %v", pyModuleName, err)
    }
    err = json.Unmarshal(out.Bytes(), &mod)
    if err != nil {
        return mod, fmt.Errorf("failed to unmarshal JSON output: %v", err)
    }
    return mod, nil
}



func generateFromConfig(cfg Config, outDir string) {
	// extract symbol message from the python module
	moduleName := pyenv.GetModuleName(cfg.LibName)		// Todo: main module and sub modules
	fmt.Printf("Extracting symbols from Python module %s...\n", moduleName)
	mod, err := pydump(moduleName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if mod.Name != moduleName {
		log.Printf("error: import module failed: %s\n", mod.Name)
		os.Exit(1)
	}

	fmt.Printf("Module %s dumped successfully, %d symbols found\n", moduleName, len(mod.Items))

	// generate Go code based on the symbols
	genGoBindings(mod, moduleName, outDir)
}


func genGoBindings(mod module, pyModuleName string, outDir string) {
	// create go package
	pkgName := getPkgName(pyModuleName)
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
	ret := types.NewTuple(pkg.NewParam(0, "", objPtr))
	
	// add context
	ctx := &context{pkg, obj, objPtr, ret, nil, nil, py}
	ctx.genMod(pkg, &mod)
	skips := ctx.skips
	if n := len(skips); n > 0 {
		log.Printf("==> Skip %d symbols:\n%v\n", n, skips)
	}

	// pkg.WriteTo(os.Stdout)
	// write to file
    outputFile := filepath.Join(outDir, pkgName+".go")
    file, err := createFileWithDirs(outputFile)
    if err != nil {
        log.Printf("error: failed to create output file %s: %v\n", outputFile, err)
        os.Exit(1)
    }
    defer file.Close()

    err = pkg.WriteTo(file)
    if err != nil {
        log.Printf("error: failed to write Go code to file %s: %v\n", outputFile, err)
        os.Exit(1)
    }
}


// go package name from Python module name
func getPkgName(pyModuleName string) string {
	if pos := strings.LastIndexByte(pyModuleName, '.'); pos >= 0 {
		return pyModuleName[pos+1:]
	}
	return pyModuleName
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
	for _, sym := range mod.Items {
		if inFuncSet(sym.Type) {
			ctx.genFunc(pkg, sym)
			continue
		}
		ctx.todos = append(ctx.todos, element{name: sym.Name, pyType: sym.Type})
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
	params, variadic := ctx.genParams(pkg, symSig)
	name = genName(name, -1)
	sig := types.NewSignatureType(nil, nil, nil, params, ctx.ret, variadic)
	fn := pkg.NewFuncDecl(token.NoPos, name, sig)
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


// 下划线转驼峰, 若为关键字则末尾添加下划线
func genName(name string, idxDontTitle int) string {
	parts := strings.Split(name, "_")
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
	return name
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

