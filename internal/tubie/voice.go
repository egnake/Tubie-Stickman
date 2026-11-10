package tubie

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ConvertReference re-encodes a given audio file to 24000Hz mono WAV.
func ConvertReference(source string, dest string) error {
	ffmpeg := BinaryPath("ffmpeg")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp(filepath.Dir(dest), ".tubie-convert-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	tmpOut := filepath.Join(tempDir, "converted.wav")
	cmd := exec.Command(ffmpeg, "-v", "error", "-y", "-i", source, "-ar", "24000", "-ac", "1", tmpOut)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg convert failed: %v: %s", err, string(out))
	}

	return os.Rename(tmpOut, dest)
}

// GetAudioDuration returns the duration of an audio/video file in seconds via ffprobe.
func GetAudioDuration(filePath string) (float64, error) {
	ffprobe := BinaryPath("ffprobe")
	cmd := exec.Command(ffprobe, "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed on %s: %v", filePath, err)
	}

	str := strings.TrimSpace(string(out))
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse audio duration '%s': %v", str, err)
	}
	return val, nil
}

// SynthesizeEdgeTTS generates speech via edge-tts CLI and converts it to 24000Hz mono WAV.
func SynthesizeEdgeTTS(text, voiceName, outWav string) (float64, error) {
	if voiceName == "" {
		voiceName = "tr-TR-AhmetNeural"
	}

	tempDir, err := os.MkdirTemp("", "tubie-tts-")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(tempDir)

	tempMp3 := filepath.Join(tempDir, "narration.mp3")

	// Determine edge-tts runner (either binary or python -m edge_tts)
	var edgeCmd *exec.Cmd
	if p, err := exec.LookPath("edge-tts"); err == nil {
		edgeCmd = exec.Command(p, "--voice", voiceName, "--rate", "+12%", "--pitch", "+0Hz", "--text", text, "--write-media", tempMp3)
	} else {
		pyExe := BinaryPath("python")
		edgeCmd = exec.Command(pyExe, "-m", "edge_tts", "--voice", voiceName, "--rate", "+12%", "--pitch", "+0Hz", "--text", text, "--write-media", tempMp3)
	}

	var stderr bytes.Buffer
	edgeCmd.Stderr = &stderr
	if err := edgeCmd.Run(); err != nil {
		return 0, fmt.Errorf("edge-tts synthesis failed: %v (details: %s)", err, stderr.String())
	}

	if err := os.MkdirAll(filepath.Dir(outWav), 0755); err != nil {
		return 0, err
	}

	ffmpeg := BinaryPath("ffmpeg")
	convCmd := exec.Command(ffmpeg, "-v", "error", "-y", "-i", tempMp3, "-ar", "24000", "-ac", "1", outWav)
	if out, err := convCmd.CombinedOutput(); err != nil {
		return 0, fmt.Errorf("ffmpeg mp3 to wav conversion failed: %v: %s", err, string(out))
	}

	return GetAudioDuration(outWav)
}

// SynthesizeVoice handles voice generation for pro, cloned, or auto modes.
func SynthesizeVoice(text string, outWav string, mode string, voiceName string, root string) (float64, string, error) {
	sesimWav := filepath.Join(root, "sesim.wav")
	modeUsed := mode
	if modeUsed == "" || modeUsed == "auto" {
		if fi, err := os.Stat(sesimWav); err == nil && !fi.IsDir() && fi.Size() > 0 {
			modeUsed = "cloned"
		} else {
			modeUsed = "pro"
		}
	}

	if modeUsed == "cloned" {
		if fi, err := os.Stat(sesimWav); err != nil || fi.IsDir() || fi.Size() == 0 {
			return 0, "", fmt.Errorf("klonlanmış ses için referans kaydı bulunamadı: %s", sesimWav)
		}
		// If XTTS is available in Python environment, invoke it
		pyCode := fmt.Sprintf(`import sys
from TTS.api import TTS
tts = TTS("tts_models/multilingual/multi-dataset/xtts_v2").to("cpu")
tts.tts_to_file(text=%q, speaker_wav=%q, language="tr", file_path=%q, speed=1.08, repetition_penalty=4.5)
`, text, sesimWav, outWav)
		cmd := exec.Command(BinaryPath("python"), "-c", pyCode)
		if out, err := cmd.CombinedOutput(); err != nil {
			return 0, "", fmt.Errorf("XTTS synthesis failed: %v: %s", err, string(out))
		}
		dur, err := GetAudioDuration(outWav)
		return dur, "cloned", err
	}

	dur, err := SynthesizeEdgeTTS(text, voiceName, outWav)
	return dur, "pro", err
}

