package pyenv

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)



// use python --version to check the version 3.12
func checkPyVersion(pyHome string) error {
	binDir := filepath.Join(pyHome, "bin")
	re := regexp.MustCompile(`^python3\.(\d+)$`)
	files, err := os.ReadDir(binDir)
	if err != nil {
		return err
	}
	version := ""
	for _, file := range files {
		if re.MatchString(file.Name()) {
			version = strings.TrimPrefix(file.Name(), "python")
			break
		}
	}
	if version == "" {	
		return fmt.Errorf("python3.12 not found")
	}
	if version != "3.12" {
		return fmt.Errorf("expect python3.12, but got %s", version)
	}
	return nil
}





