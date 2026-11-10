//go:build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type wordTiming struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Word  string  `json:"word"`
}

type whisperWord struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Word  string  `json:"word"`
}

type whisperSegment struct {
	Words []whisperWord `json:"words"`
}

type whisperOutput struct {
	Segments []whisperSegment `json:"segments"`
}

func main() {
	outputFlag := flag.String("output", "", "Output path for aligned words JSON")
	modelFlag := flag.String("model", "base", "Whisper model name")
	languageFlag := flag.String("language", "", "Spoken language code")
	_ = flag.String("device", "cpu", "Execution device (cpu or cuda)")
	allowDownloadFlag := flag.Bool("allow-model-download", false, "Allow automatic model download")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: align_captions AUDIO --output WORDS.json [--model base] [--language en] [--allow-model-download]\n")
		os.Exit(2)
	}

	audioPath := flag.Arg(0)
	if *outputFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: --output is required\n")
		os.Exit(2)
	}

	if _, err := os.Stat(*outputFlag); err == nil {
		fmt.Fprintf(os.Stderr, "Output exists; will not overwrite\n")
		os.Exit(2)
	}

	if fi, err := os.Stat(audioPath); err != nil || fi.IsDir() {
		fmt.Fprintf(os.Stderr, "Audio file not found\n")
		os.Exit(2)
	}

	// Check if standalone whisper or whisper-cli is available on PATH
	var whisperTool string
	for _, tool := range []string{"whisper", "whisper-cli"} {
		if path, err := exec.LookPath(tool); err == nil {
			whisperTool = path
			break
		}
	}

	if whisperTool == "" {
		fmt.Fprintf(os.Stderr, "Optional dependency missing: install whisper or whisper-cli, or provide existing aligned captions\n")
		os.Exit(2)
	}

	// Check cache
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			cacheDir = filepath.Join(home, ".cache")
		}
	}
	cacheFile := filepath.Join(cacheDir, "whisper", fmt.Sprintf("%s.pt", *modelFlag))
	if _, err := os.Stat(cacheFile); err != nil && !*allowDownloadFlag {
		fmt.Fprintf(os.Stderr, "Model not cached; --allow-model-download is required for a download\n")
		os.Exit(2)
	}

	tempDir, err := os.MkdirTemp("", "align-captions-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	args := []string{audioPath, "--model", *modelFlag, "--output_format", "json", "--output_dir", tempDir}
	if *languageFlag != "" {
		args = append(args, "--language", *languageFlag)
	}

	cmd := exec.Command(whisperTool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}

	baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	jsonPath := filepath.Join(tempDir, baseName+".json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read whisper output: %v\n", err)
		os.Exit(1)
	}

	var parsed whisperOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse whisper output: %v\n", err)
		os.Exit(1)
	}

	var words []wordTiming
	for _, seg := range parsed.Segments {
		for _, w := range seg.Words {
			if w.End > w.Start {
				words = append(words, wordTiming{
					Start: w.Start,
					End:   w.End,
					Word:  w.Word,
				})
			}
		}
	}

	if len(words) == 0 {
		fmt.Fprintf(os.Stderr, "No aligned words returned; inspect the audio\n")
		os.Exit(1)
	}

	outPayload := map[string]any{
		"words":  words,
		"method": "whisper word_timestamps; human text/timing review required",
	}
	encoded, err := json.MarshalIndent(outPayload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode result: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFlag, encoded, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
		os.Exit(1)
	}

	summary, _ := json.Marshal(map[string]any{
		"output": *outputFlag,
		"words":  len(words),
	})
	fmt.Println(string(summary))
}
