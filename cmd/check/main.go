package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// Fallback to executable location
	exe, err := os.Executable()
	if err == nil {
		dir = filepath.Dir(exe)
		for {
			if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", fmt.Errorf("could not find repository root containing go.mod")
}

func runStep(name string, dir string, exe string, args ...string) error {
	fmt.Printf("\nRunning: %s %s (in %s)\n", exe, strings.Join(args, " "), dir)
	cmd := exec.Command(exe, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func main() {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	npmExe, err := exec.LookPath("npm")
	if err != nil {
		npmExe = "npm"
	}

	steps := []struct {
		name string
		dir  string
		exe  string
		args []string
	}{
		{
			name: "Internal Tubie tests",
			dir:  root,
			exe:  "go",
			args: []string{"test", "-v", "./internal/tubie/..."},
		},
		{
			name: "Root verification tests",
			dir:  root,
			exe:  "go",
			args: []string{"test", "-v", "./tests/..."},
		},
		{
			name: "Directing stickman skill tests",
			dir:  filepath.Join(root, "skills", "directing-stickman-videos"),
			exe:  "go",
			args: []string{"test", "-v", "./..."},
		},
		{
			name: "Node npm tests",
			dir:  root,
			exe:  npmExe,
			args: []string{"test"},
		},
	}

	for _, step := range steps {
		if err := runStep(step.name, step.dir, step.exe, step.args...); err != nil {
			fmt.Fprintf(os.Stderr, "\nStep %q failed: %v\n", step.name, err)
			os.Exit(1)
		}
	}

	fmt.Println("\nAll checks passed.")
}
