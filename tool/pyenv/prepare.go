package pyenv

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Prepare() {
	pyHome := os.Getenv("PYTHONHOME")
	if pyHome != "" {
		err := checkPyVersion(pyHome)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("use python from ", pyHome)
		return
	}
	// system default python
	names := []string{"python-3.12", "python-3.12-embed", "python3"}
	for _, name := range names {
		cmd := exec.Command("pkg-config", "--variable=libdir", name)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		libPath := strings.TrimSpace(string(out))
		pyHome = filepath.Join(libPath, "../")
		err = checkPyVersion(pyHome)
		if err != nil {
			continue
		}
		os.Setenv("PYTHONHOME", pyHome)
		break
	}
	pyHome = os.Getenv("PYTHONHOME")
	if pyHome == "" {
		log.Fatal("python3.12 not found")
	}
	fmt.Println("use python from: ", pyHome)
}