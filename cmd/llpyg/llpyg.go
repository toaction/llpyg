package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"encoding/json"
	"path/filepath"
)

// libName to module name mapping
var libMainModule = map[string]string{
    "scikit-learn": "sklearn",
    "pillow":       "PIL",
}

type Config struct {
	Name 		string `json:"name"`
	LibName 	string `json:"libName"`
	LibVersion 	string `json:"libVersion"`
}


type symbol struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
	Sig  string `json:"sig"`
}

type module struct {
	Name  string    `json:"name"`
	Items []*symbol `json:"items"`
}

var funcSet []string

func init() {
	funcSet = []string{
		"function", "method", "builtin_function_or_method",
	}
	funcSet = append(funcSet, "ufunc")
	funcSet = append(funcSet, "method-wrapper")
	funcSet = append(funcSet, "_ArrayFunctionDispatcher")
}


func inFuncSet(typeName string) bool {
	for _, item := range funcSet {
		if item == typeName {
			return true
		}
	}
	return false
}


// get the main module name for the library
func getMainModuleName(libName string) string {
    if mod, ok := libMainModule[libName]; ok {
        return mod
    }
    return libName // default to the library name itself
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


// install Python library using pip
func installPythonLib(libName, libVersion string) bool {
	arg := libName
	if libVersion != "" {
		arg += "==" + libVersion
	}
    installCmd := "pip3"
    args := []string{"install", arg}
	// system packages may not be writable
	args = append(args, "--break-system-packages")
    cmd := exec.Command(installCmd, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to install library %s\n", arg)
		return false
	}
	return true
}


func getVersionFromPip(libName string) (string, error) {
	cmd := exec.Command("pip3", "show", libName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error: failed to get version for library %s", libName)
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Version: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Version: ")), nil
		}
	}
	return "", fmt.Errorf("error: version not found for library %s", libName)
}


// 检查本地是否存在Python环境, 是否已安装指定版本的Python库
func pyEnvCheck(libName, libVersion string) bool {
	// check python env
	cmd := exec.Command("python3", "--version")
    if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error: Python is not installed or not found")
        return false
    }
	// check python lib
	cmd = exec.Command("pip3", "show", libName)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: library %s is not installed\n", libName)
		// prompt to install the library
		fmt.Printf("Do you want to install the library %s? (y/n): ", libName)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" {
			return false
		}
		res := installPythonLib(libName, libVersion)
		if !res {
			// installation failed
			return false
		}
	}
	// check python lib version
	if libVersion == "" {
		return true
	}
	version, err := getVersionFromPip(libName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	// compare version
	if version != libVersion {
		fmt.Fprintf(os.Stderr, "error: library %s version mismatch, expected %s, got %s\n", libName, libVersion, version)
		fmt.Fprintf(os.Stderr, "Do you want to install the library %s with version %s? (y/n): ", libName, libVersion)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" {
			installed := installPythonLib(libName, libVersion)
			if !installed {
				fmt.Fprintln(os.Stderr, "error: failed to install the library with specified version")
			}
		}
	}
	return true
}


func createFileWithDirs(filePath string) (*os.File, error) {
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, err
    }
    return os.Create(filePath)
}


func main() {
	// python library name and version, main module name
	// e.g. scikit-learn==1.0.2 : scikit-learn, 1.0.2, sklearn
	var libName, libVersion, moduleName string

	output := flag.String("output", "./test", "Output dir")
	flag.Parse()
	// fmt.Println("Output dir:", *output)

	// get library name and version from command line argument
    if flag.NArg() < 1 {
        fmt.Fprintln(os.Stderr, "Usage: lpyg [pythonLibName[==version]] [-o outputDir]")
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
	envChecked := pyEnvCheck(libName, libVersion)
	if !envChecked {
		os.Exit(1)
	}

	// get library name and version
	
	version, err := getVersionFromPip(libName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Python library (%s version %s) is ready\n", libName, version)

	// generate llpyg.cfg
	moduleName = getMainModuleName(libName)
	cfg := Config{
		Name: 		strings.ReplaceAll(moduleName, "-", "_"),
		LibName: 	libName,
		LibVersion: version,
	}
	outDir := filepath.Join(*output, cfg.Name)
	cfgPath := filepath.Join(outDir, "llpyg.cfg")
    file, err := createFileWithDirs(cfgPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer file.Close()
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "error: failed to write JSON to file: %v\n", err)
        os.Exit(1)
    }
	fmt.Printf("Configuration file created at %s\n", cfgPath)

	generateFromConfig(cfg, outDir)
}


func generateFromConfig(cfg Config, outDir string) {
	// extract symbol message from the library(main module)
	moduleName := getMainModuleName(cfg.LibName)
	mod, err := pydump(moduleName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to dump Python module %s: %v\n", moduleName, err)
		os.Exit(1)
	}
	dumpToJson(mod, outDir)

}
