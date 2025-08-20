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
	// check python library
	if err := checkLibrary(pycmd, libName); err != nil {
		log.Fatal(err)
	}
}

func checkPython() (pycmd string, err error) {
	var version string
	// try python3
	cmd := exec.Command("python3", "--version")
	output, err := cmd.Output()
	if err == nil {
		version = string(output)
		if strings.HasPrefix(version, "Python 3.12") {
			return "python3", nil
		}
	}
	// try python
	cmd = exec.Command("python", "--version")
	output, err = cmd.Output()
	if err == nil {
		version = string(output)
		if strings.HasPrefix(version, "Python 3.12") {
			return "python", nil
		}
	}
	if version != "" {
		return "", fmt.Errorf("error: Python version is not 3.12: %s", version)
	}
	return "", fmt.Errorf("error: Python is not installed or not found")
}

// Python library name to module name mapping
var libToModule = map[string]string{
	"scikit-learn": "sklearn",
	"pillow":       "PIL",
}

func GetModuleName(libName string) string {
	if mod, ok := libToModule[libName]; ok {
		return mod
	}
	return libName
}

func checkLibrary(pycmd, libName string) error {
	moduleName := GetModuleName(libName)
	code := fmt.Sprintf("import %s; print(%s.__version__)", moduleName, moduleName)
	cmd := exec.Command(pycmd, "-c", code)
	var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error: %s check failed: %s", libName, stderr.String())
	}
	fmt.Printf("%s %s is ready\n", libName, strings.TrimSpace(stdout.String()))
	return nil
}
