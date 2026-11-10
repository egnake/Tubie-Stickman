package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"directing-stickman-videos/pkg/a6"
)

func TestA6FinishOfflineRenderer(t *testing.T) {
	ffmpeg := os.Getenv("FFMPEG")
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	if _, err := exec.LookPath(ffmpeg); err != nil {
		t.Skip("ffmpeg not found, skipping video synthesis test")
	}

	temp := t.TempDir()
	picturePath := filepath.Join(temp, "picture.mp4")
	mixPath := filepath.Join(temp, "mix.m4a")

	// 1. Generate 2s 360x640 test picture
	cmdV := exec.Command(ffmpeg, "-v", "error", "-f", "lavfi", "-i",
		"color=c=0x382638:s=360x640:r=24:d=2", "-c:v", "libx264",
		"-pix_fmt", "yuv420p", picturePath)
	if err := cmdV.Run(); err != nil {
		t.Fatalf("Failed to create test video with ffmpeg: %v", err)
	}

	// 2. Generate 2s 440Hz test AAC audio
	cmdA := exec.Command(ffmpeg, "-v", "error", "-f", "lavfi", "-i",
		"sine=frequency=440:duration=2", "-c:a", "aac", mixPath)
	if err := cmdA.Run(); err != nil {
		t.Fatalf("Failed to create test audio with ffmpeg: %v", err)
	}

	cfg := map[string]any{
		"picture": "picture.mp4",
		"audio":   "mix.m4a",
		"script":  "Go go now.",
		"title":   []string{"WHAT IF", "TIME STOPPED?"},
		"cues": []map[string]any{
			{"start": 0.0, "end": 2.0, "text": "GO GO NOW."},
		},
		"words": []map[string]any{
			{"start": 0.0, "end": 0.5, "word": "Go"},
			{"start": 0.5, "end": 1.0, "word": "go"},
			{"start": 1.0, "end": 1.7, "word": "now."},
		},
	}

	configPath := filepath.Join(temp, "config.json")
	writeConfig := func(c map[string]any) {
		b, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}
		if err := os.WriteFile(configPath, b, 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
	}

	writeConfig(cfg)

	// Validate-only check
	valResult, err := a6.Render(configPath, "", true)
	if err != nil {
		t.Fatalf("Render validate_only failed: %v", err)
	}
	if validated, ok := valResult["validated"].(bool); !ok || !validated {
		t.Fatalf("Expected validated: true, got %v", valResult["validated"])
	}

	// Negative time checks for words and cues
	for _, field := range []string{"words", "cues"} {
		invalid := clonePlan(cfg)
		seq := invalid[field].([]any)
		first := seq[0].(map[string]any)
		first["end"] = -1.0
		writeConfig(invalid)
		if _, err := a6.Render(configPath, "", true); err == nil {
			t.Fatalf("Expected invalid %s to fail validation, but succeeded", field)
		}
	}

	// Mismatched script tokens check
	invalidScript := clonePlan(cfg)
	invalidScript["script"] = "Missing words."
	writeConfig(invalidScript)
	if _, err := a6.Render(configPath, "", true); err == nil {
		t.Fatal("Expected mismatched narration script to fail validation, but succeeded")
	}

	// Reset config to valid
	writeConfig(cfg)

	// Repeated GO highlight position check
	skillRoot := findSkillRoot(t)
	fontPath := filepath.Join(skillRoot, "assets", "fonts", "NotoSans.ttf")
	firstLayer, err := a6.CaptionLayer("GO GO NOW.", 0, 360, 640, fontPath)
	if err != nil {
		t.Fatalf("CaptionLayer activeIndex=0 failed: %v", err)
	}
	secondLayer, err := a6.CaptionLayer("GO GO NOW.", 1, 360, 640, fontPath)
	if err != nil {
		t.Fatalf("CaptionLayer activeIndex=1 failed: %v", err)
	}

	amber1 := make(map[int]bool)
	for y := 0; y < firstLayer.Bounds().Dy(); y++ {
		for x := 0; x < 360; x++ {
			c := firstLayer.NRGBAAt(x, y)
			if c.R == 255 && c.G == 203 && c.B == 74 {
				amber1[x] = true
			}
		}
	}

	amber2 := make(map[int]bool)
	for y := 0; y < secondLayer.Bounds().Dy(); y++ {
		for x := 0; x < 360; x++ {
			c := secondLayer.NRGBAAt(x, y)
			if c.R == 255 && c.G == 203 && c.B == 74 {
				amber2[x] = true
			}
		}
	}

	if len(amber1) == 0 {
		t.Fatal("Expected amber pixels in first layer (activeIndex=0), found none")
	}
	if len(amber2) == 0 {
		t.Fatal("Expected amber pixels in second layer (activeIndex=1), found none")
	}

	maxAmber1 := -1
	for x := range amber1 {
		if x > maxAmber1 {
			maxAmber1 = x
		}
	}
	minAmber2 := 999999
	for x := range amber2 {
		if x < minAmber2 {
			minAmber2 = x
		}
	}

	if !(maxAmber1 < minAmber2) {
		t.Fatalf("Positional highlight failed: max(amber1)=%d is not < min(amber2)=%d", maxAmber1, minAmber2)
	}

	// Full Render test
	finalDir := filepath.Join(temp, "final")
	renderResult, err := a6.Render(configPath, finalDir, false)
	if err != nil {
		t.Fatalf("Render to %s failed: %v", finalDir, err)
	}

	decodePass, _ := renderResult["decode_pass"].(bool)
	audioCopied, _ := renderResult["audio_packets_copied"].(bool)
	if !decodePass || !audioCopied {
		t.Fatalf("Expected decode_pass=true and audio_packets_copied=true, got decode_pass=%v, audio_packets_copied=%v",
			decodePass, audioCopied)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(finalDir, "manifest.json"))
	if err != nil {
		t.Fatalf("Failed to read manifest.json: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("Failed to parse manifest.json: %v", err)
	}

	if manifest["creator_watermark"] != false {
		t.Fatalf("Expected creator_watermark=false, got %v", manifest["creator_watermark"])
	}
	if manifest["generation_calls"] != float64(0) {
		t.Fatalf("Expected generation_calls=0, got %v", manifest["generation_calls"])
	}

	srtBytes, err := os.ReadFile(filepath.Join(finalDir, "captions.srt"))
	if err != nil {
		t.Fatalf("Failed to read captions.srt: %v", err)
	}
	if !strings.Contains(string(srtBytes), "GO GO NOW.") {
		t.Fatalf("captions.srt does not contain 'GO GO NOW.': %s", string(srtBytes))
	}

	// No-overwrite check on non-empty directory
	if _, err := a6.Render(configPath, finalDir, false); err == nil {
		t.Fatal("Expected render on existing non-empty output directory to fail, but succeeded")
	}
}
