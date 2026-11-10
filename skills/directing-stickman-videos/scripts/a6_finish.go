//go:build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"directing-stickman-videos/pkg/a6"
)

func main() {
	configFlag := flag.String("config", "", "Path to A6 finishing JSON configuration")
	outputDirFlag := flag.String("output-dir", "", "Path to output directory")
	validateOnlyFlag := flag.Bool("validate-only", false, "Validate configuration and media without rendering")
	flag.Parse()

	if *configFlag == "" {
		fmt.Fprintf(os.Stderr, "A6 finishing failed: --config is required\n")
		os.Exit(1)
	}

	result, err := a6.Render(*configFlag, *outputDirFlag, *validateOnlyFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "A6 finishing failed: %v\n", err)
		os.Exit(1)
	}

	formatted, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "A6 finishing failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(formatted))
}
