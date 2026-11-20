package tubie

import (
	"bytes"
	"context"
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

// BuildViralVocalFilter constructs the FFmpeg broadcast mastering chain:
// silence tightening (zero dead air at 0:00, crisp inter-sentence pacing), studio EQ, compression, and -15 LUFS normalization.
func BuildViralVocalFilter(atempo float64) string {
	parts := make([]string, 0, 7)
	if atempo > 0 && atempo != 1.0 {
		parts = append(parts, fmt.Sprintf("atempo=%.2f", atempo))
	}
	parts = append(parts,
		"silenceremove=start_periods=1:start_silence=0.02:start_threshold=-42dB:stop_periods=-1:stop_duration=0.28:stop_silence=0.14:stop_threshold=-42dB",
		"highpass=f=75",
		"equalizer=f=145:t=q:w=1.2:g=2.0",
		"equalizer=f=3200:t=q:w=1.4:g=2.4",
		"acompressor=threshold=-16dB:ratio=3.5:attack=10:release=60",
		"loudnorm=I=-15:TP=-1.5:LRA=9",
	)
	return strings.Join(parts, ",")
}

// SynthesizeEdgeTTS generates speech via edge-tts CLI, applies viral mastering, and converts to 24kHz mono WAV.
func SynthesizeEdgeTTS(text, voiceName, rate, pitch, outWav string) (float64, error) {
	return SynthesizeEdgeTTSContext(context.Background(), text, voiceName, rate, pitch, outWav)
}

// SynthesizeEdgeTTSContext synthesizes speech via edge-tts CLI with cancellation support.
func SynthesizeEdgeTTSContext(ctx context.Context, text, voiceName, rate, pitch, outWav string) (float64, error) {
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
		edgeCmd = exec.CommandContext(ctx, p, "--voice", voiceName, "--rate", rate, "--pitch", pitch, "--text", text, "--write-media", tempMp3)
	} else {
		pyExe := BinaryPath("python")
		edgeCmd = exec.CommandContext(ctx, pyExe, "-m", "edge_tts", "--voice", voiceName, "--rate", rate, "--pitch", pitch, "--text", text, "--write-media", tempMp3)
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
	convCmd := exec.CommandContext(ctx, ffmpeg, "-v", "error", "-y", "-i", tempMp3,
		"-af", BuildViralVocalFilter(1.0),
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
	return BuildViralVocalFilter(atempo)
}

func normalizeSpeechText(text string) string {
	replacer := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ")
	fields := strings.Fields(replacer.Replace(text))
	return strings.Join(fields, " ")
}

// xttsScript reads UTF-8 JSON arguments from stdin (eliminating code string interpolation).
const xttsScript = `import json, sys
args = json.loads(sys.stdin.buffer.read().decode("utf-8"))
from TTS.api import TTS
tts = TTS("tts_models/multilingual/multi-dataset/xtts_v2").to("cpu")
tts.tts_to_file(
    text=args["text"],
    speaker_wav=args["speaker_wav"],
    language="tr",
    file_path=args["file_path"],
    speed=float(args["speed"]),
    repetition_penalty=4.5,
)
`

// SynthesizeVoice handles voice generation for pro, cloned, or auto modes with custom voice styling.
func SynthesizeVoice(text string, outWav string, mode string, voiceStyle string, root string) (float64, string, error) {
	return SynthesizeVoiceContext(context.Background(), text, outWav, mode, voiceStyle, root)
}

// SynthesizeVoiceContext synthesizes narration with context cancellation and mastering.
func SynthesizeVoiceContext(ctx context.Context, text string, outWav string, mode string, voiceStyle string, root string) (float64, string, error) {
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

		payload, err := json.Marshal(map[string]any{
			"text":        normalizeSpeechText(text),
			"speaker_wav": sesimWav,
			"file_path":   rawClonedWav,
			"speed":       xttsSpeed,
		})
		if err != nil {
			return 0, "", err
		}

		cmd := exec.CommandContext(ctx, BinaryPath("python"), "-c", xttsScript)
		cmd.Stdin = bytes.NewReader(payload)
		if out, err := cmd.CombinedOutput(); err != nil {
			return 0, "", fmt.Errorf("XTTS synthesis failed: %v: %s", err, string(out))
		}

		// Apply pitch-preserving speedup + silence tightening + broadcast EQ and loudnorm
		ffmpeg := BinaryPath("ffmpeg")
		convCmd := exec.CommandContext(ctx, ffmpeg, "-v", "error", "-y", "-i", rawClonedWav,
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

	dur, err := SynthesizeEdgeTTSContext(ctx, text, voiceName, rate, pitch, outWav)
	return dur, "pro", err
}
