package pyenv

import (
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
	var pipcmd string
	switch pycmd {
	case "python3":
		pipcmd = "pip3"
	case "python":
		pipcmd = "pip"
	}
	if err := checkLibrary(pipcmd, libName); err != nil {
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

func checkLibrary(pipcmd, libName string) error {
	cmd := exec.Command(pipcmd, "show", libName)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error: %s is not installed: %v", libName, err)
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Version: ") {
			fmt.Printf("%s %s is ready\n", libName, strings.TrimPrefix(line, "Version: "))
			break
		}
	}
	return nil
}