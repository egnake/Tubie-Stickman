package tubie

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ThumbnailOptions configures aspect ratio, A/B visual variant, and mood for generated covers.
type ThumbnailOptions struct {
	Aspect  string
	Variant string
	Mood    string
}

// GenerateDedicatedThumbnail produces a click-worthy, high-impact vertical cover image.
func GenerateDedicatedThumbnail(root string, title string, topic string, propPath string, poseIndex int, outPath string, filmMp4 string) error {
	return GenerateDedicatedThumbnailWithOptions(root, title, topic, propPath, poseIndex, ThumbnailOptions{
		Aspect:  "9:16",
		Variant: "a",
		Mood:    "danger",
	}, outPath, filmMp4)
}

// GenerateDedicatedThumbnailWithOptions produces a 9:16 or 16:9 viral cyber cover in variant A or B.
func GenerateDedicatedThumbnailWithOptions(root string, title string, topic string, propPath string, poseIndex int, opts ThumbnailOptions, outPath string, filmMp4 string) error {
	nodeExe := BinaryPath("node")
	scriptPath := filepath.Join(root, "internal", "tubie", "thumbnail.cjs")

	aspect := strings.TrimSpace(opts.Aspect)
	if aspect != "16:9" {
		aspect = "9:16"
	}
	variant := strings.ToLower(strings.TrimSpace(opts.Variant))
	if variant != "b" {
		variant = "a"
	}
	mood := strings.TrimSpace(opts.Mood)
	if mood == "" {
		mood = "danger"
	}

	args := []string{
		scriptPath,
		"--title", title,
		"--topic", topic,
		"--pose", strconv.Itoa(poseIndex),
		"--aspect", aspect,
		"--variant", variant,
		"--mood", mood,
		"--out", outPath,
	}
	if propPath != "" {
		args = append(args, "--prop", propPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, nodeExe, args...)
	cmd.Dir = root
	cmd.Env = os.Environ()

	if out, err := cmd.CombinedOutput(); err == nil && isFile(outPath) {
		return nil
	} else {
		fmt.Printf("⚠️ Dedicated thumbnail script failed (%v: %s), falling back to high-res video frame\n", err, string(out))
	}

	if filmMp4 != "" && isFile(filmMp4) {
		ffmpegExe := BinaryPath("ffmpeg")
		ffCtx, ffCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer ffCancel()
		fallbackCmd := exec.CommandContext(ffCtx, ffmpegExe, "-y", "-ss", "00:00:03.500", "-i", filmMp4, "-frames:v", "1", "-q:v", "2", outPath)
		if err := fallbackCmd.Run(); err == nil && isFile(outPath) {
			return nil
		}
	}

	return fmt.Errorf("failed to generate thumbnail at %s", outPath)
}
