package main

import (
	"os"

	"github.com/goplus/llpyg/tool/pyenv"
)

func main() {
	envTest()
}


func libTest() {
	pyenv.PyEnvCheck("numpy")		// installed
	// pyenv.PyEnvCheck("pillow")		// uninstalled
	// pyenv.PyEnvCheck("scikit-learn")	// require scipy not installed
}

func envTest() {
	// export PYTHONHOME=/Users/mac/work/python
	// export PATH=$PYTHONHOME/bin:$PATH
	os.Setenv("PATH", "/Users/mac/work/python/bin:$PATH")
	libTest()
}