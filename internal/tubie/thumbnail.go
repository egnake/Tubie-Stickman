package tubie

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// GenerateDedicatedThumbnail produces a click-worthy, high-impact vertical cover image.
func GenerateDedicatedThumbnail(root string, title string, topic string, propPath string, poseIndex int, outPath string, filmMp4 string) error {
	nodeExe := BinaryPath("node")
	scriptPath := filepath.Join(root, "internal", "tubie", "thumbnail.cjs")

	args := []string{
		scriptPath,
		"--title", title,
		"--topic", topic,
		"--pose", strconv.Itoa(poseIndex),
		"--out", outPath,
	}
	if propPath != "" {
		args = append(args, "--prop", propPath)
	}

	cmd := exec.Command(nodeExe, args...)
	cmd.Dir = root
	cmd.Env = os.Environ()

	if out, err := cmd.CombinedOutput(); err == nil && isFile(outPath) {
		return nil
	} else {
		// Log warning and fallback
		fmt.Printf("⚠️ Dedicated thumbnail script failed (%v: %s), falling back to high-res video frame\n", err, string(out))
	}

	// Fallback to high-clarity video frame at 3.50s (when title and character are fully rendered)
	if filmMp4 != "" && isFile(filmMp4) {
		ffmpegExe := BinaryPath("ffmpeg")
		fallbackCmd := exec.Command(ffmpegExe, "-y", "-ss", "00:00:03.500", "-i", filmMp4, "-frames:v", "1", "-q:v", "2", outPath)
		if err := fallbackCmd.Run(); err == nil && isFile(outPath) {
			return nil
		}
	}

	return fmt.Errorf("failed to generate thumbnail at %s", outPath)
}
