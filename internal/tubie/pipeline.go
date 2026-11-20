package tubie

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Reporter sends structured progress and stage updates.
type Reporter func(event map[string]any)

// VideoFormatSpec defines output resolution, aspect ratio, and caption placement preset.
type VideoFormatSpec struct {
	Name        string `json:"name"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Aspect      string `json:"aspect"`
	CaptionY    int    `json:"caption_y"`
	CaptionSize int    `json:"caption_size"`
}

// ResolveVideoFormat resolves canonical video aspect/resolution presets for vertical 9:16 and horizontal 16:9 exports.
func ResolveVideoFormat(raw string) VideoFormatSpec {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "shorts-hd", "9:16-hd", "1080x1920":
		return VideoFormatSpec{Name: "shorts-hd", Width: 1080, Height: 1920, Aspect: "9:16", CaptionY: 1780, CaptionSize: 56}
	case "youtube", "16:9", "1280x720", "landscape":
		return VideoFormatSpec{Name: "youtube", Width: 1280, Height: 720, Aspect: "16:9", CaptionY: 630, CaptionSize: 44}
	case "youtube-hd", "16:9-hd", "1920x1080":
		return VideoFormatSpec{Name: "youtube-hd", Width: 1920, Height: 1080, Aspect: "16:9", CaptionY: 630, CaptionSize: 44}
	default:
		return VideoFormatSpec{Name: "shorts", Width: 720, Height: 1280, Aspect: "9:16", CaptionY: 1780, CaptionSize: 56}
	}
}

// Pipeline orchestrates synthesis, storyboard rendering, and export.
type Pipeline struct {
	Root       string
	Mode       string
	VoiceStyle string
	BGM        string
	Format     string
	LoopEnding bool
	Reporter   Reporter
	Review     func(projectDir string) error
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
	return p.ProduceContext(context.Background(), preset)
}

// ProduceContext executes the complete video production with cancellation support and multi-format export.
func (p *Pipeline) ProduceContext(ctx context.Context, preset Preset) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	fmtSpec := ResolveVideoFormat(p.Format)

	fmt.Printf("\n=================================================================\n")
	fmt.Printf("🎬 SİBER GÜVENLİK VİRAL VİDEO MOTORU: %s [%s %dx%d]\n", preset.Title, fmtSpec.Aspect, fmtSpec.Width, fmtSpec.Height)
	fmt.Printf("=================================================================\n")

	outDir := filepath.Join(p.Root, "output")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	slug := sanitizeSlug(preset.Slug, preset.Title)
	preset.Slug = slug

	projectDir := filepath.Join(p.Root, "projects", preset.Slug)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", err
	}

	if len(preset.Scenes) == 0 {
		preset.Scenes = EnsureScenesForScript(preset.Title, preset.Narration)
	}

	// 1. Synthesize Audio
	if p.Reporter != nil {
		p.Reporter(map[string]any{"stage": "voice", "label": "Seslendirme hazırlanıyor", "slug": preset.Slug, "format": fmtSpec.Name})
	}
	narrationWav := filepath.Join(projectDir, "narration.wav")
	dur, modeUsed, err := SynthesizeVoiceContext(ctx, preset.Narration, narrationWav, p.Mode, p.VoiceStyle, p.Root)
	if err != nil {
		return "", fmt.Errorf("seslendirme hatası: %w", err)
	}
	fmt.Printf("✅ Seslendirme tamamlandı! (%s) Süre: %.3f saniye\n", strings.ToUpper(modeUsed), dur)

	if err := ctx.Err(); err != nil {
		return "", err
	}

	// 2. Generate Acoustically Synced Viral Captions, Storyboard Shots & Sound Design
	if p.Reporter != nil {
		p.Reporter(map[string]any{"stage": "captions", "label": "Kusursuz altyazılar ve sahneler senkronize ediliyor", "slug": preset.Slug})
	}

	captions, timedWords := AlignAndBuildCaptions(preset.Narration, narrationWav, dur)
	boundaries := ComputeSceneBoundaries(preset.Scenes, timedWords, preset.Narration, dur)
	shots := BuildStoryboardWithOptions(preset.Scenes, dur, StoryboardOptions{
		Width:      fmtSpec.Width,
		Height:     fmtSpec.Height,
		TimedWords: timedWords,
		Narration:  preset.Narration,
	})
	sfxList := BuildSoundDesignWithTimings(preset.Scenes, boundaries, dur)

	projectConfig := map[string]any{
		"version":            1,
		"style":              "B1",
		"theme":              "cyber",
		"backgroundColor":    "#0A0E17",
		"width":              fmtSpec.Width,
		"height":             fmtSpec.Height,
		"fps":                30,
		"duration":           dur,
		"captionStyle":       "viral",
		"captionSize":        fmtSpec.CaptionSize,
		"captionY":           fmtSpec.CaptionY,
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

	// Dynamic Background Music (BGM)
	bgmChoice := strings.ToLower(strings.TrimSpace(p.BGM))
	if bgmChoice == "" {
		bgmChoice = "cyber"
	}
	if bgmChoice != "none" {
		trackFile := fmt.Sprintf("bgm_%s.wav", bgmChoice)
		bgmPath := filepath.Join(p.Root, "assets", "audio", trackFile)
		if _, err := os.Stat(bgmPath); err != nil {
			bgmPath = filepath.Join(p.Root, "assets", "audio", "bgm_cyber.wav")
		}
		if _, err := os.Stat(bgmPath); err == nil {
			_ = copyFile(bgmPath, filepath.Join(projectDir, "bgm.wav"))
			projectConfig["music"] = map[string]any{"path": "bgm.wav", "gain": 0.16}
			projectConfig["bgm"] = map[string]any{"path": "bgm.wav", "gain": 0.16}
		}
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
		if err := ctx.Err(); err != nil {
			return err
		}
		args := []string{autopilotScript, subcommand, projectConfigFile}
		args = append(args, extraArgs...)
		cmd := exec.CommandContext(ctx, nodeExe, args...)
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
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			if p.Reporter != nil {
				trimmed := strings.TrimSpace(strings.TrimPrefix(line, "TUBIE_EVENT "))
				var ev map[string]any
				if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
					if err := json.Unmarshal([]byte(trimmed), &ev); err == nil {
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
		p.Reporter(map[string]any{"stage": "export", "label": "Video paketi, altyazılar, viral analiz ve kapaklar kaydediliyor", "slug": preset.Slug})
	}

	videoFolder := filepath.Join(outDir, preset.Slug)
	if err := os.MkdirAll(videoFolder, 0755); err != nil {
		return "", err
	}

	filmMp4 := filepath.Join(projectDir, "output", "film.mp4")
	finalMp4 := filepath.Join(videoFolder, preset.Slug+".mp4")
	if err := copyFile(filmMp4, finalMp4); err != nil {
		return "", fmt.Errorf("film.mp4 export failed: %w", err)
	}

	// 1. SubRip (.srt) & WebVTT (.vtt) subtitle exports
	cleanSrt := FormatSRT(captions)
	_ = os.WriteFile(filepath.Join(videoFolder, preset.Slug+".srt"), []byte(cleanSrt), 0644)
	_ = os.WriteFile(filepath.Join(projectDir, "output", "film.srt"), []byte(cleanSrt), 0644)

	cleanVtt := FormatVTT(captions)
	_ = os.WriteFile(filepath.Join(videoFolder, preset.Slug+".vtt"), []byte(cleanVtt), 0644)
	_ = os.WriteFile(filepath.Join(projectDir, "output", "film.vtt"), []byte(cleanVtt), 0644)

	// 2. aciklama.txt (A/B/C/D/E test titles, pinned comment, viral hashtags)
	viralMeta := GenerateViralMetadata(preset.Title, preset.Narration)
	viralText := FormatViralText(preset.Title, preset.Narration, viralMeta)
	_ = os.WriteFile(filepath.Join(videoFolder, "aciklama.txt"), []byte(viralText), 0644)

	// 3. viral_report.json
	viralReport := AnalyzeScriptVirality(preset.Title, preset.Narration, preset.Scenes, p.LoopEnding)
	reportPayload := map[string]any{
		"total_score":       viralReport.TotalScore,
		"grade":             viralReport.Grade,
		"estimated_seconds": viralReport.EstimatedSeconds,
		"word_count":        viralReport.WordCount,
		"hook_sentence":     viralReport.HookSentence,
		"threat_category":   viralReport.ThreatCategory,
		"loop_ready":        viralReport.LoopReady,
		"dimensions":        viralReport.Dimensions,
		"tips":              viralReport.Tips,
		"format":            fmtSpec.Name,
		"aspect":            fmtSpec.Aspect,
		"width":             fmtSpec.Width,
		"height":            fmtSpec.Height,
	}
	if reportBytes, err := json.MarshalIndent(reportPayload, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(videoFolder, "viral_report.json"), reportBytes, 0644)
	}

	// 4. High-CTR dedicated thumbnails (9:16 A, 9:16 B, 16:9 A)
	thumbMood := "danger"
	if len(preset.Scenes) > 0 && strings.TrimSpace(preset.Scenes[0].Mood) != "" {
		thumbMood = preset.Scenes[0].Mood
	}
	propAsset := ""
	if len(preset.Scenes) > 0 && preset.Scenes[0].Prop != nil {
		propAsset = preset.Scenes[0].Prop.Path
	}
	thumbSpecs := []struct {
		file    string
		aspect  string
		variant string
		pose    int
	}{
		{"thumbnail.jpg", "9:16", "a", 2},
		{"thumbnail_b.jpg", "9:16", "b", 1},
		{"thumbnail_16x9.jpg", "16:9", "a", 2},
	}
	for _, spec := range thumbSpecs {
		outThumb := filepath.Join(videoFolder, spec.file)
		if err := GenerateDedicatedThumbnailWithOptions(p.Root, preset.Title, preset.Narration, propAsset, spec.pose, ThumbnailOptions{
			Aspect:  spec.aspect,
			Variant: spec.variant,
			Mood:    thumbMood,
		}, outThumb, filmMp4); err != nil {
			fmt.Printf("⚠️ Thumbnail oluşturulamadı (%s): %v\n", spec.file, err)
		}
	}

	fmt.Printf("\n✨ Başarıyla tamamlandı! Video paketi klasörü: %s\n", videoFolder)
	return finalMp4, nil
}

func sanitizeSlug(raw, fallbackTitle string) string {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	var sb strings.Builder
	for _, r := range cleaned {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-':
			sb.WriteRune(r)
		case r == ' ':
			sb.WriteByte('_')
		}
	}
	res := strings.Trim(sb.String(), "._-")
	if res == "" {
		res = slugify(fallbackTitle)
	}
	if res == "" {
		res = "siber_guvenlik_video"
	}
	return res
}
