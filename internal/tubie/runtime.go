package tubie

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FindRoot locates the repository root containing go.mod.
func FindRoot() (string, error) {
	// First check current working directory
	if dir, err := os.Getwd(); err == nil {
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

	// Fallback to executable location
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
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

	return "", fmt.Errorf("repository root containing go.mod not found")
}

// BinaryPath finds an executable on PATH, with system and WinGet links as fallback.
func BinaryPath(name string) string {
	override := os.Getenv(strings.ToUpper(name))
	if override != "" {
		return override
	}

	// For ffmpeg/ffprobe on Windows, prioritize C:\ffmpeg\bin if present
	if runtime.GOOS == "windows" && (name == "ffmpeg" || name == "ffprobe") {
		preferred := filepath.Join("C:\\ffmpeg\\bin", name+".exe")
		if fi, err := os.Stat(preferred); err == nil && !fi.IsDir() {
			return preferred
		}
	}

	if path, err := exec.LookPath(name); err == nil {
		return path
	}

	if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			candidate := filepath.Join(localAppData, "Microsoft", "WinGet", "Links", name+".exe")
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				return candidate
			}
		}
	}

	return name
}

// CUDADevice returns "cuda" if nvidia-smi runs successfully, else "cpu".
func CUDADevice() string {
	cmd := exec.Command("nvidia-smi")
	if err := cmd.Run(); err == nil {
		return "cuda"
	}
	return "cpu"
}

