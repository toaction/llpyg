package pyenv

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func PyEnvCheck(libName string) {
	// check python env
	pycmd, err := checkPython()
	if err != nil {
		log.Fatal(err)
	}
	pyHome, err := pyLocation(pycmd)
	if err == nil {
		fmt.Printf("Python home: %s\n", pyHome)
	}
	// check python library
	if err := checkLibrary(pycmd, libName); err != nil {
		log.Fatal(err)
	}
}

func checkPython() (pycmd string, err error) {
	var version string
	// check python
	cmd := exec.Command("python3", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error: python3 is not installed or not found")
		// Todo: support python, python3.12(>= 3.12)
	}
	// check version: >= 3.12
	version = strings.TrimSpace(strings.TrimPrefix(string(output), "Python "))
	parts := strings.Split(version, ".")
	if len(parts) < 2 || parts[0] != "3" || parts[1] < "12" {
		return "", fmt.Errorf("error: Python version does not match, should be >= 3.12, found: %s", version)
	}
	return "python3", nil
}

func pyLocation(pycmd string) (pyHome string, err error) {
	cmd := exec.Command(pycmd, "-c", "import sys; print(sys.prefix)")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Python library name to module name mapping
var libToModule = map[string]string{
	"scikit-learn": "sklearn",
	"pillow":       "PIL",
}

func getModuleName(libName string) string {
	if mod, ok := libToModule[libName]; ok {
		return mod
	}
	return libName
}

func checkLibrary(pycmd, libName string) error {
	// check library
	pipcmd := strings.Replace(pycmd, "python", "pip", 1)
	cmd := exec.Command(pipcmd, "show", libName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: %s is not installed", libName)
	}
	// check module
	moduleName := getModuleName(libName)
	code := fmt.Sprintf("import %s; print(%s.__version__)", moduleName, moduleName)
	cmd = exec.Command(pycmd, "-c", code)
	var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error: %s import module failed: %s", libName, stderr.String())
	}
	fmt.Printf("%s %s is ready\n", libName, strings.TrimSpace(stdout.String()))
	return nil
}

