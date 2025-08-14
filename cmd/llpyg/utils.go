package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"log"
	"os/exec"
)


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

func createFileWithDirs(filePath string) (*os.File, error) {
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, err
    }
    return os.Create(filePath)
}


func writeConfig(cfg Config, outDir string) error {
	cfgPath := filepath.Join(outDir, "llpyg.cfg")
    file, err := createFileWithDirs(cfgPath)
    if err != nil {
        return fmt.Errorf("error: failed to create file: %w", err)
    }
    defer file.Close()
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(cfg); err != nil {
        return fmt.Errorf("error: failed to write configuration: %w", err)
    }
	return nil
}


func genGoModule(modName string, outDir string) {
    if err := os.Chdir(outDir); err != nil {
		log.Printf("error: failed to change directory to %s: %v\n", outDir, err)
        os.Exit(1)
    }
    // init go module
    cmd := exec.Command("go", "mod", "init", modName)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        log.Printf("error: failed to initialize Go module: %v\n", err)
        os.Exit(1)
    }
    // go get github.com/goplus/lib/py
    getCmd := exec.Command("go", "get", "github.com/goplus/lib/py")
    getCmd.Stdout = os.Stdout
    getCmd.Stderr = os.Stderr
    if err := getCmd.Run(); err != nil {
        log.Printf("error: failed to get github.com/goplus/lib/py: %v\n", err)
        os.Exit(1)
    }

    // go mod tidy
    // tidyCmd := exec.Command("go", "mod", "tidy")
    // tidyCmd.Stdout = os.Stdout
    // tidyCmd.Stderr = os.Stderr
    // if err := tidyCmd.Run(); err != nil {
    //     log.Printf("error: failed to run go mod tidy: %v\n", err)
	// 	os.Exit(1)
    // }
}