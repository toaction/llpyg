package pyenv

import (
	"os"
	"runtime"
)

func Prepare() {
	pyHome := os.Getenv("PYTHONHOME")
	if pyHome == "" {		// use system
		return
	}
	// bin
	binPath := os.Getenv("PATH")
	os.Setenv("PATH", pyHome+"/bin:"+binPath)
	// pkg_config_path
	pkgConfigPath := os.Getenv("PKG_CONFIG_PATH")
	os.Setenv("PKG_CONFIG_PATH", pyHome+"/lib/pkgconfig:"+pkgConfigPath)
	// lib
	switch runtime.GOOS {
	case "darwin":
		libPath := os.Getenv("DYLD_LIBRARY_PATH")
		os.Setenv("DYLD_LIBRARY_PATH", pyHome+"/lib:"+libPath)
	case "linux":
		libPath := os.Getenv("LD_LIBRARY_PATH")
		os.Setenv("LD_LIBRARY_PATH", pyHome+"/lib:"+libPath)
	}
}