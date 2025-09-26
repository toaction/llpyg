package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		runMode  string
		wantArgs Args
	}{
		{
			name:    "basic_cmd_mode",
			args:    []string{"numpy"},
			runMode: "cmd",
			wantArgs: Args{
				OutputDir: "./out",
				ModName:   "",
				ModDepth:  1,
				Kwarg:     "numpy",
			},
		},
		{
			name:    "cmd_mode_with_options",
			args:    []string{"-o", "test", "-mod", "numpy", "-d", "2", "numpy"},
			runMode: "cmd",
			wantArgs: Args{
				OutputDir: "test",
				ModName:   "numpy",
				ModDepth:  2,
				Kwarg:     "numpy",
			},
		},
		{
			name:    "cfg_mode",
			args:    []string{"-o", "test", "-mod", "numpy", "llpyg.cfg"},
			runMode: "cfg",
			wantArgs: Args{
				OutputDir: "test",
				ModName:   "numpy",
				ModDepth:  1,
				Kwarg:     "llpyg.cfg",
			},
		},
		{
			name:    "default_values",
			args:    []string{"pandas"},
			runMode: "cmd",
			wantArgs: Args{
				OutputDir: "./out",
				ModName:   "",
				ModDepth:  1,
				Kwarg:     "pandas",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			originalArgs := os.Args
			defer func() {
				os.Args = originalArgs
			}()
			os.Args = append([]string{"llpyg"}, c.args...)
			runMode, args := parseArgs()
			if runMode != c.runMode {
				t.Errorf("runMode = %v, want %v", runMode, c.runMode)
			}
			if !equalArgsWithPath(args, c.wantArgs) {
				t.Errorf("args = %v, want %v", args, c.wantArgs)
			}
		})
	}
}

func equalArgsWithPath(a, b Args) bool {
	aAbs, err := filepath.Abs(a.OutputDir)
	if err != nil {
		return false
	}
	bAbs, err := filepath.Abs(b.OutputDir)
	if err != nil {
		return false
	}

	if aAbs != bAbs {
		return false
	}
	if a.ModName != b.ModName {
		return false
	}
	if a.ModDepth != b.ModDepth {
		return false
	}
	if a.Kwarg != b.Kwarg {
		return false
	}
	return true
}
