package pyenv

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)


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
		log.Printf("error: failed to install library %s: %v\n", arg, err)
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
// return result, installed version
func PyEnvCheck(libName, libVersion string) (res bool, version string) {
	// check python env
	cmd := exec.Command("python3", "--version")
    if err := cmd.Run(); err != nil {
		log.Printf("error: Python is not installed or not found: %v\n", err)
        return false,""
    }
	// check python lib
	cmd = exec.Command("pip3", "show", libName)
	if err := cmd.Run(); err != nil {
		log.Printf("error: library %s is not installed\n", libName)
		// prompt to install the library
		fmt.Printf("Do you want to install the library %s? (y/n): ", libName)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" {
			return false, ""
		}
		res := installPythonLib(libName, libVersion)
		if !res {
			// installation failed
			return false, ""
		}
	}
	// check python lib version
	version, err := getVersionFromPip(libName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false, ""
	}
	if libVersion == "" {
		return true, version
	}
	// compare version
	if version != libVersion {
		log.Printf("error: library %s version mismatch, expected %s, got %s\n", libName, libVersion, version)
		fmt.Printf("Do you want to install the library %s with version %s? (y/n): ", libName, libVersion)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" {
			installed := installPythonLib(libName, libVersion)
			if !installed {
				// installation failed, return exist version
				return true, version
			}
		}
	}
	return true, libVersion
}