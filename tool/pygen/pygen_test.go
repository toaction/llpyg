package pygen

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"path/filepath"
)

func readJsonData(t *testing.T, path string) module {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var mod module
	err = json.Unmarshal(data, &mod)
	if err != nil {
		t.Fatal(err)
	}
	return mod
}

func TestGenFunc(t *testing.T) {
	mod := readJsonData(t, "testdata/func/demo.json")
	ctx := createGoPackage(mod)
	for _, sym := range mod.Functions {
		ctx.genFunc(ctx.pkg, sym)
	}
	err := compareWithExpected(t, ctx, "testdata/func/expect.go")
	if err != nil {
		t.Fatalf("test gen func failed: %v", err)
	}
	t.Logf("test gen func pass")
}

func compareWithExpected(t *testing.T, ctx *context, expectedPath string) error {
	outFilePath := "./temp/actual_git.go"
	dir := filepath.Dir(outFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	ctx.pkg.WriteTo(outFile)
	cmd := exec.Command("go", "fmt", outFilePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go fmt failed: %v", err)
	}
	cmd = exec.Command("git", "diff", "--no-index", outFilePath, expectedPath)
	diffOutput, err := cmd.CombinedOutput()
	if err != nil {
		if len(diffOutput) > 0 {
			return fmt.Errorf("get not match expected: \n%s", string(diffOutput))
		}
		t.Fatal(err)
	}
	return nil
}
