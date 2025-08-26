package main

import (
	"os"

	"github.com/goplus/llpyg/tool/pyenv"
)

func main() {
	envTest()
}


func libTest() {
	// pyenv.PyEnvCheck("numpy")		// installed
	pyenv.PyEnvCheck("pillow")		// uninstalled
	// pyenv.PyEnvCheck("scikit-learn")	// require scipy not installed
}

func envTest() {
	os.Setenv("PYTHONHOME", "/Users/mac/work/xgo/python")
	pyenv.Prepare()
	libTest()
}

