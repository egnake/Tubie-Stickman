package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type previewTarget struct {
	stem      string
	extension string
	fps       int
	colors    int
	scale     int
	width     int
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "assets", "readme")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// Try from executable location
	exe, err := os.Executable()
	if err == nil {
		dir = filepath.Dir(exe)
		for {
			if _, err := os.Stat(filepath.Join(dir, "assets", "readme")); err == nil {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ".", nil
}

func main() {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ffmpeg is required.")
		os.Exit(1)
	}

	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find repo root: %v\n", err)
		os.Exit(1)
	}

	media := filepath.Join(root, "assets", "readme")
	previewsDir := filepath.Join(media, "previews")
	if err := os.MkdirAll(previewsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create previews dir: %v\n", err)
		os.Exit(1)
	}

	originals := []string{
		"light-theme-demo", "dark-theme-demo", "style-2a-demo", "style-2b-demo",
		"style-3-demo", "style-3-community-fire", "style-4-cozy-station", "style-4-cozy-cafe",
	}

	var targets []previewTarget
	for _, stem := range originals {
		targets = append(targets, previewTarget{stem, "gif", 8, 96, 4, 400})
	}
	targets = append(targets,
		previewTarget{"style-5-japan", "mp4", 6, 64, 5, 400},
		previewTarget{"style-5-egypt", "mp4", 6, 64, 5, 400},
		previewTarget{"style-5-japan", "mp4", 6, 64, 5, 640},
	)

	for _, t := range targets {
		output := fmt.Sprintf("%s.gif", t.stem)
		if t.width == 640 {
			output = fmt.Sprintf("%s-hero.gif", t.stem)
		}
		height := t.width * 9 / 16
		filters := fmt.Sprintf("[0:v]fps=%d,scale=%d:%d:flags=lanczos,split[a][b];[a]palettegen=max_colors=%d:stats_mode=diff[p];[b][p]paletteuse=dither=bayer:bayer_scale=%d:diff_mode=rectangle",
			t.fps, t.width, height, t.colors, t.scale)

		inputPath := filepath.Join(media, fmt.Sprintf("%s.%s", t.stem, t.extension))
		outputPath := filepath.Join(previewsDir, output)

		args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", inputPath}
		if t.extension == "mp4" {
			args = append(args, "-t", "10")
		}
		args = append(args, "-filter_complex", filters, "-loop", "0", outputPath)

		cmd := exec.Command(ffmpeg, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed optimizing %s: %v\n", t.stem, err)
			os.Exit(1)
		}
	}

	fmt.Println("Built lightweight README previews; original media unchanged.")
}
