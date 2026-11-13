package tubie

import (
	"bytes"
	"encoding/json"
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

// SynthesizeEdgeTTS generates speech via edge-tts CLI, applies vocal compression/loudnorm,
// and converts it to 24000Hz mono WAV.
func SynthesizeEdgeTTS(text, voiceName, rate, pitch, outWav string) (float64, error) {
	if voiceName == "" {
		voiceName = "tr-TR-AhmetNeural"
	}
	if rate == "" {
		rate = "+15%"
	}
	if pitch == "" {
		pitch = "+0Hz"
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
		edgeCmd = exec.Command(p, "--voice", voiceName, "--rate", rate, "--pitch", pitch, "--text", text, "--write-media", tempMp3)
	} else {
		pyExe := BinaryPath("python")
		edgeCmd = exec.Command(pyExe, "-m", "edge_tts", "--voice", voiceName, "--rate", rate, "--pitch", pitch, "--text", text, "--write-media", tempMp3)
	}

	var stderr bytes.Buffer
	edgeCmd.Stderr = &stderr
	if err := edgeCmd.Run(); err != nil {
		return 0, fmt.Errorf("edge-tts synthesis failed: %v (details: %s)", err, stderr.String())
	}

	if err := os.MkdirAll(filepath.Dir(outWav), 0755); err != nil {
		return 0, err
	}

	// High-impact broadcast vocal chain: dynamic compression + loudness normalization to -16 LUFS
	ffmpeg := BinaryPath("ffmpeg")
	convCmd := exec.Command(ffmpeg, "-v", "error", "-y", "-i", tempMp3,
		"-af", "acompressor=threshold=-16dB:ratio=3.5:attack=10:release=60,loudnorm=I=-16:TP=-1.5:LRA=10",
		"-ar", "24000", "-ac", "1", outWav)
	if out, err := convCmd.CombinedOutput(); err != nil {
		return 0, fmt.Errorf("ffmpeg mp3 to wav conversion failed: %v: %s", err, string(out))
	}

	return GetAudioDuration(outWav)
}

// ClonedVoiceTiming returns the XTTS synthesis speed and pitch-preserving FFmpeg atempo multiplier.
// Keeping XTTS speed in a crisp range while applying FFmpeg atempo yields energetic pacing without slurring.
func ClonedVoiceTiming(voiceStyle string) (xttsSpeed float64, atempo float64) {
	switch strings.TrimSpace(voiceStyle) {
	case "cloned-fast":
		return 1.16, 1.18
	default:
		return 1.15, 1.10
	}
}

// BuildClonedAudioFilter constructs the FFmpeg audio filter chain for cloned voice polishing and tempo.
func BuildClonedAudioFilter(atempo float64) string {
	base := "acompressor=threshold=-16dB:ratio=3.5:attack=10:release=60,loudnorm=I=-16:TP=-1.5:LRA=10"
	if atempo > 0 && atempo != 1.0 {
		return fmt.Sprintf("atempo=%.2f,%s", atempo, base)
	}
	return base
}

func normalizeSpeechText(text string) string {
	replacer := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ")
	fields := strings.Fields(replacer.Replace(text))
	return strings.Join(fields, " ")
}

// SynthesizeVoice handles voice generation for pro, cloned, or auto modes with custom voice styling.
func SynthesizeVoice(text string, outWav string, mode string, voiceStyle string, root string) (float64, string, error) {
	sesimWav := filepath.Join(root, "sesim.wav")
	modeUsed := mode

	// If mode is auto or empty, default to high-retention pro viral voice
	if modeUsed == "" || modeUsed == "auto" {
		modeUsed = "pro"
	}

	if modeUsed == "cloned" || strings.HasPrefix(strings.TrimSpace(voiceStyle), "cloned") {
		if fi, err := os.Stat(sesimWav); err != nil || fi.IsDir() || fi.Size() == 0 {
			return 0, "", fmt.Errorf("klonlanmış ses için referans kaydı bulunamadı: %s", sesimWav)
		}

		if err := os.MkdirAll(filepath.Dir(outWav), 0755); err != nil {
			return 0, "", err
		}

		xttsSpeed, atempo := ClonedVoiceTiming(voiceStyle)
		rawClonedWav := filepath.Join(filepath.Dir(outWav), "raw_cloned.wav")
		defer os.Remove(rawClonedWav)

		textJSON, _ := json.Marshal(normalizeSpeechText(text))
		speakerJSON, _ := json.Marshal(sesimWav)
		outJSON, _ := json.Marshal(rawClonedWav)

		pyCode := fmt.Sprintf(`import json
from TTS.api import TTS
tts = TTS("tts_models/multilingual/multi-dataset/xtts_v2").to("cpu")
tts.tts_to_file(
    text=json.loads(%q),
    speaker_wav=json.loads(%q),
    language="tr",
    file_path=json.loads(%q),
    speed=%.2f,
    repetition_penalty=4.5,
)
`, string(textJSON), string(speakerJSON), string(outJSON), xttsSpeed)

		cmd := exec.Command(BinaryPath("python"), "-c", pyCode)
		if out, err := cmd.CombinedOutput(); err != nil {
			return 0, "", fmt.Errorf("XTTS synthesis failed: %v: %s", err, string(out))
		}

		// Apply pitch-preserving speedup + broadcast compression and loudnorm
		ffmpeg := BinaryPath("ffmpeg")
		convCmd := exec.Command(ffmpeg, "-v", "error", "-y", "-i", rawClonedWav,
			"-af", BuildClonedAudioFilter(atempo),
			"-ar", "24000", "-ac", "1", outWav)
		if out, err := convCmd.CombinedOutput(); err != nil {
			return 0, "", fmt.Errorf("ffmpeg cloned voice filter failed: %v: %s", err, string(out))
		}

		dur, err := GetAudioDuration(outWav)
		return dur, "cloned", err
	}

	// Resolve voice parameters for viral pro mode
	voiceName := "tr-TR-AhmetNeural"
	rate := "+15%"
	pitch := "+0Hz"

	switch voiceStyle {
	case "viral-fast":
		rate = "+22%"
	case "viral-female":
		voiceName = "tr-TR-EmelNeural"
		rate = "+14%"
	case "deep-male":
		rate = "+8%"
		pitch = "-4Hz"
	default: // "viral-male", "pro", or ""
		rate = "+15%"
		pitch = "+0Hz"
	}

	dur, err := SynthesizeEdgeTTS(text, voiceName, rate, pitch, outWav)
	return dur, "pro", err
}
