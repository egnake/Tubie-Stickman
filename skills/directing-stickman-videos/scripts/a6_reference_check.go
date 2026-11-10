//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"directing-stickman-videos/pkg/a6"
)

func findDefaultRoot() string {
	candidates := []string{".", "..", filepath.Join("..", "..")}
	for _, c := range candidates {
		if fi, err := os.Stat(filepath.Join(c, "assets", "a6", "reference-manifest.json")); err == nil && !fi.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}
	return "."
}

func main() {
	rootFlag := flag.String("root", "", "Path to installed Skill directory")
	planFlag := flag.String("plan", "", "Path to reference plan JSON")
	projRootFlag := flag.String("project-root", "", "Path to project root for project references")
	showPrompts := flag.Bool("show-prompts", false, "Show expanded image prompts")
	flag.Parse()

	root := *rootFlag
	if root == "" {
		root = findDefaultRoot()
	}

	assets, err := a6.ValidatePack(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	var expanded []map[string]any
	if *planFlag != "" {
		expanded, err = a6.ValidatePlanWithProject(root, *planFlag, *projRootFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("PASS: %d A6 reference files; %d image prompts checked.\n", len(assets), len(expanded))
	if *showPrompts {
		for _, item := range expanded {
			fmt.Printf("\n%s\n%s\n", item["id"], item["prompt"])
		}
	}
	fmt.Println("Offline integrity/structure only; visual quality, tool attachments and generation not verified.")
}
