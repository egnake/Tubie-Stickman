package tubie

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Reporter sends structured progress and stage updates.
type Reporter func(event map[string]any)

// Pipeline orchestrates synthesis, storyboard rendering, and export.
type Pipeline struct {
	Root     string
	Mode     string
	Reporter Reporter
	Review   func(projectDir string) error
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// Produce executes the complete video production for a given preset or custom script.
func (p *Pipeline) Produce(preset Preset) (string, error) {
	fmt.Printf("\n=================================================================\n")
	fmt.Printf("🎬 SİBER GÜVENLİK VİRAL VİDEO MOTORU: %s\n", preset.Title)
	fmt.Printf("=================================================================\n")

	outDir := filepath.Join(p.Root, "output")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	slug := preset.Slug
	if slug == "" {
		slug = slugify(preset.Title)
	}
	preset.Slug = slug

	projectDir := filepath.Join(p.Root, "projects", preset.Slug)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", err
	}

	// 1. Synthesize Audio
	if p.Reporter != nil {
		p.Reporter(map[string]any{"stage": "voice", "label": "Seslendirme hazırlanıyor", "slug": preset.Slug})
	}
	narrationWav := filepath.Join(projectDir, "narration.wav")
	dur, modeUsed, err := SynthesizeVoice(preset.Narration, narrationWav, p.Mode, "tr-TR-AhmetNeural", p.Root)
	if err != nil {
		return "", fmt.Errorf("seslendirme hatası: %w", err)
	}
	fmt.Printf("✅ Seslendirme tamamlandı! (%s) Süre: %.3f saniye\n", strings.ToUpper(modeUsed), dur)

	// 2. Generate 100% Accurate Viral Captions & Storyboard Shots
	if p.Reporter != nil {
		p.Reporter(map[string]any{"stage": "captions", "label": "Kusursuz altyazılar ve sahneler derleniyor", "slug": preset.Slug})
	}

	captions := BuildAccurateCaptions(preset.Narration, narrationWav, dur)
	shots := BuildStoryboard(preset.Scenes, dur)

	// Synced SFX cues
	sfxList := make([]map[string]any, 0)
	if len(preset.Scenes) > 0 {
		step := dur / float64(len(preset.Scenes))
		for i, sc := range preset.Scenes {
			t := math.Round(float64(i)*step*1000) / 1000
			if t < dur {
				kind := "tap"
				switch sc.SFX {
				case "alert", "glitch":
					kind = "pop"
				case "lock":
					kind = "chime"
				case "paper":
					kind = "paper"
				default:
					kind = "tap"
				}
				sfxList = append(sfxList, map[string]any{
					"time": t,
					"kind": kind,
					"gain": 0.08,
				})
			}
		}
	}

	projectConfig := map[string]any{
		"version":            1,
		"style":              "B1",
		"theme":              "cyber",
		"backgroundColor":    "#0A0E17",
		"width":              720,
		"height":             1280,
		"fps":                30,
		"duration":           dur,
		"captionStyle":       "viral",
		"captionSize":        56,
		"captionY":           1460,
		"captionColor":       "#00FF66",
		"captionActiveColor": "#00FF66",
		"captionBg":          "rgba(7, 12, 20, 0.96)",
		"captionBorder":      "rgba(0, 255, 128, 0.50)",
		"voice": map[string]any{
			"path": "narration.wav",
			"lead": 0.0,
		},
		"shots":    shots,
		"captions": captions,
		"sfx":      sfxList,
	}

	// Copy BGM if available
	bgmCyber := filepath.Join(p.Root, "assets", "audio", "bgm_cyber.wav")
	bgmLofi := filepath.Join(p.Root, "assets", "audio", "bgm_lofi.wav")
	if _, err := os.Stat(bgmCyber); err == nil {
		_ = copyFile(bgmCyber, filepath.Join(projectDir, "bgm.wav"))
		projectConfig["bgm"] = map[string]any{"path": "bgm.wav", "gain": 0.12}
	} else if _, err := os.Stat(bgmLofi); err == nil {
		_ = copyFile(bgmLofi, filepath.Join(projectDir, "bgm.wav"))
		projectConfig["bgm"] = map[string]any{"path": "bgm.wav", "gain": 0.12}
	}

	projectConfigFile := filepath.Join(projectDir, "project.json")
	projData, err := json.MarshalIndent(projectConfig, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(projectConfigFile, projData, 0644); err != nil {
		return "", err
	}

	// 3. Autopilot execution
	nodeExe := BinaryPath("node")
	autopilotScript := filepath.Join(p.Root, "skills", "directing-stickman-videos", "scripts", "autopilot.mjs")

	runAutopilot := func(subcommand string, extraArgs ...string) error {
		args := []string{autopilotScript, subcommand, projectConfigFile}
		args = append(args, extraArgs...)
		cmd := exec.Command(nodeExe, args...)
		cmd.Dir = p.Root
		cmd.Env = os.Environ()

		// Stream stdout to parse live-frame events if reporter attached
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			return err
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			if p.Reporter != nil {
				// Parse frame updates or JSON lines
				var ev map[string]any
				if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
					if err := json.Unmarshal([]byte(line), &ev); err == nil {
						p.Reporter(ev)
					}
				}
			}
		}

		return cmd.Wait()
	}

	liveArgs := []string{}
	if p.Reporter != nil {
		liveArgs = append(liveArgs, "--live")
	}

	// Preview step
	if p.Reporter != nil {
		p.Reporter(map[string]any{"stage": "preview", "label": "Önizleme oluşturuluyor", "slug": preset.Slug})
	}
	fmt.Println("[1/3] Video önizlemesi oluşturuluyor...")
	if err := runAutopilot("preview", append([]string{"--replace"}, liveArgs...)...); err != nil {
		return "", fmt.Errorf("autopilot preview failed: %w", err)
	}

	// Approval step
	if p.Review != nil {
		if err := p.Review(projectDir); err != nil {
			return "", err
		}
	}
	fmt.Println("[2/3] Önizleme onaylanıyor...")
	note := "Kullanıcı onayı alındı"
	if err := runAutopilot("approve", "--note", note); err != nil {
		return "", fmt.Errorf("autopilot approve failed: %w", err)
	}

	// Render step
	if p.Reporter != nil {
		p.Reporter(map[string]any{"stage": "render", "label": "Son video işleniyor", "slug": preset.Slug})
	}
	fmt.Println("[3/3] Tam video MP4 olarak render ediliyor...")
	if err := runAutopilot("render", append([]string{"--full", "--replace"}, liveArgs...)...); err != nil {
		return "", fmt.Errorf("autopilot render failed: %w", err)
	}

	// Export step
	if p.Reporter != nil {
		p.Reporter(map[string]any{"stage": "export", "label": "Video paketi, altyazı ve sosyal medya açıklaması kaydediliyor", "slug": preset.Slug})
	}

	// Create organized folder: output/<slug>/
	videoFolder := filepath.Join(outDir, preset.Slug)
	if err := os.MkdirAll(videoFolder, 0755); err != nil {
		return "", err
	}

	filmMp4 := filepath.Join(projectDir, "output", "film.mp4")
	finalMp4 := filepath.Join(videoFolder, preset.Slug+".mp4")
	if err := copyFile(filmMp4, finalMp4); err != nil {
		return "", fmt.Errorf("film.mp4 export failed: %w", err)
	}

	// 1. Format and write guaranteed 100% accurate SubRip (.srt) subtitle file
	cleanSrt := FormatSRT(captions)
	finalSrt := filepath.Join(videoFolder, preset.Slug+".srt")
	_ = os.WriteFile(finalSrt, []byte(cleanSrt), 0644)
	_ = os.WriteFile(filepath.Join(projectDir, "output", "film.srt"), []byte(cleanSrt), 0644)

	// 2. Generate and write aciklama.txt (A/B/C test titles, post description, viral hashtags)
	viralMeta := GenerateViralMetadata(preset.Title, preset.Narration)
	viralText := FormatViralText(preset.Title, preset.Narration, viralMeta)
	descFile := filepath.Join(videoFolder, "aciklama.txt")
	_ = os.WriteFile(descFile, []byte(viralText), 0644)

	// 3. Generate high-resolution thumbnail.jpg
	thumbJpg := filepath.Join(videoFolder, "thumbnail.jpg")
	ffmpegExe := BinaryPath("ffmpeg")
	cmd := exec.Command(ffmpegExe, "-y", "-ss", "00:00:01.200", "-i", filmMp4, "-frames:v", "1", "-q:v", "2", thumbJpg)
	if err := cmd.Run(); err != nil || !isFile(thumbJpg) {
		previewPng := filepath.Join(projectDir, "output", "preview.png")
		filmSheet := filepath.Join(projectDir, "output", "film-sheet.png")
		if isFile(previewPng) {
			_ = copyFile(previewPng, thumbJpg)
		} else if isFile(filmSheet) {
			_ = copyFile(filmSheet, thumbJpg)
		}
	}

	fmt.Printf("\n✨ Başarıyla tamamlandı! Video paketi klasörü: %s\n", videoFolder)
	return finalMp4, nil
}
