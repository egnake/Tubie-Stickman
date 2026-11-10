package a6

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	tokenRegex = regexp.MustCompile(`[A-Za-z0-9]+(?:'[A-Za-z0-9]+)*`)
)

type WordCue struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Word  string  `json:"word"`
}

type CaptionCue struct {
	Start float64   `json:"start"`
	End   float64   `json:"end"`
	Text  string    `json:"text"`
	Words []WordCue `json:"-"`
}

type FinishConfig struct {
	Picture     string       `json:"picture"`
	Audio       string       `json:"audio"`
	Script      string       `json:"script"`
	Title       any          `json:"title"`
	Cues        []CaptionCue `json:"cues"`
	Words       []WordCue    `json:"words"`
	CaptionFont string       `json:"caption_font,omitempty"`
	TitleFont   string       `json:"title_font,omitempty"`
}

type PreparedMedia struct {
	Config      *FinishConfig
	Picture     string
	Audio       string
	CaptionFont string
	TitleFont   string
	Width       int
	Height      int
	FPS         float64
	FPSStr      string
	Duration    float64
	Sound       map[string]any
}

// Tokens splits text into uppercase alphanumeric words with optional apostrophes.
func Tokens(text string) []string {
	normalized := strings.ReplaceAll(strings.ToUpper(text), "’", "'")
	return tokenRegex.FindAllString(normalized, -1)
}

// Stamp formats seconds as SRT timestamp 00:00:00,000.
func Stamp(t float64) string {
	ms := int64(math.Round(t * 1000.0))
	if ms < 0 {
		ms = 0
	}
	hours := ms / 3600000
	minutes := (ms / 60000) % 60
	seconds := (ms / 1000) % 60
	millis := ms % 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, millis)
}

// Sha computes the SHA-256 hex digest of a file.
func Sha(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	buf := make([]byte, 1024*1024)
	if _, err := io.CopyBuffer(h, f, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// AudioHash streams AAC packets from an audio file and computes their SHA-256 digest.
func AudioHash(path string) (string, error) {
	ffmpeg := os.Getenv("FFMPEG")
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	cmd := exec.Command(ffmpeg, "-v", "error", "-i", path, "-map", "0:a:0", "-c", "copy", "-f", "adts", "-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	h := sha256.New()
	buf := make([]byte, 1024*1024)
	if _, err := io.CopyBuffer(h, stdout, buf); err != nil {
		_ = cmd.Wait()
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		return "", errors.New("Unable to hash AAC packets")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Probe runs ffprobe on path and parses JSON output.
func Probe(path string) (map[string]any, error) {
	ffprobe := os.Getenv("FFPROBE")
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	out, err := exec.Command(ffprobe, "-v", "error", "-show_streams", "-show_format", "-of", "json", path).Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed on %s: %w", path, err)
	}
	var res map[string]any
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func parseFraction(str string) (float64, error) {
	str = strings.TrimSpace(str)
	if parts := strings.Split(str, "/"); len(parts) == 2 {
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil && den > 0 {
			return num / den, nil
		}
	}
	return strconv.ParseFloat(str, 64)
}

func validateSequence[T any](items []T, getTimes func(T) (float64, float64), duration float64, label string) error {
	previous := 0.0
	for _, item := range items {
		start, end := getTimes(item)
		if math.IsNaN(start) || math.IsInf(start, 0) || math.IsNaN(end) || math.IsInf(end, 0) ||
			!(0 <= start && start < end && end <= duration+0.05 && start >= previous-1e-6) {
			return fmt.Errorf("%s: invalid, overlapping or out-of-range times: %v", label, item)
		}
		previous = end
	}
	return nil
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func parseTitleLines(title any) ([]string, error) {
	var lines []string
	switch t := title.(type) {
	case string:
		for _, s := range strings.Split(t, "\n") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				lines = append(lines, trimmed)
			}
		}
	case []string:
		for _, s := range t {
			if strings.TrimSpace(s) == "" || strings.Contains(s, "\n") {
				return nil, errors.New("title must contain one or two nonempty lines")
			}
			lines = append(lines, strings.TrimSpace(s))
		}
	case []any:
		for _, item := range t {
			str, ok := item.(string)
			if !ok || strings.TrimSpace(str) == "" || strings.Contains(str, "\n") {
				return nil, errors.New("title must contain one or two nonempty lines")
			}
			lines = append(lines, strings.TrimSpace(str))
		}
	default:
		return nil, errors.New("title must contain one or two nonempty lines")
	}

	if len(lines) < 1 || len(lines) > 2 {
		return nil, errors.New("title must contain one or two nonempty lines")
	}
	return lines, nil
}

// Prepare reads and preflights the finishing configuration and referenced media.
func Prepare(configPath string) (*FinishConfig, *PreparedMedia, error) {
	absCfg, err := filepath.Abs(configPath)
	if err != nil {
		return nil, nil, err
	}
	data, err := os.ReadFile(absCfg)
	if err != nil {
		return nil, nil, err
	}
	var cfg FinishConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, nil, err
	}

	ffmpeg := os.Getenv("FFMPEG")
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	ffprobe := os.Getenv("FFPROBE")
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	for _, tool := range []string{ffmpeg, ffprobe} {
		if _, err := exec.LookPath(tool); err != nil {
			return nil, nil, fmt.Errorf("%s is required; no automatic install", tool)
		}
	}

	cfgDir := filepath.Dir(absCfg)
	resolvePath := func(val, key string) (string, error) {
		if strings.TrimSpace(val) == "" {
			return "", fmt.Errorf("Missing %s", key)
		}
		p := val
		if !filepath.IsAbs(p) {
			p = filepath.Join(cfgDir, p)
		}
		p, err = filepath.Abs(p)
		if err != nil {
			return "", err
		}
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() {
			return "", fmt.Errorf("%s not available: %s", key, p)
		}
		return p, nil
	}

	picture, err := resolvePath(cfg.Picture, "picture")
	if err != nil {
		return nil, nil, err
	}
	audio, err := resolvePath(cfg.Audio, "audio")
	if err != nil {
		return nil, nil, err
	}

	vp, err := Probe(picture)
	if err != nil {
		return nil, nil, err
	}
	ap, err := Probe(audio)
	if err != nil {
		return nil, nil, err
	}

	var videoStream, soundStream map[string]any
	if streams, ok := vp["streams"].([]any); ok {
		for _, s := range streams {
			if smap, ok := s.(map[string]any); ok && smap["codec_type"] == "video" {
				videoStream = smap
				break
			}
		}
	}
	if streams, ok := ap["streams"].([]any); ok {
		for _, s := range streams {
			if smap, ok := s.(map[string]any); ok && smap["codec_type"] == "audio" {
				soundStream = smap
				break
			}
		}
	}

	if videoStream == nil {
		return nil, nil, errors.New("Missing video stream in picture")
	}
	if soundStream == nil {
		return nil, nil, errors.New("Missing audio stream in sound")
	}

	avgFPSStr, _ := videoStream["avg_frame_rate"].(string)
	rFPSStr, _ := videoStream["r_frame_rate"].(string)
	avgFPS, err := parseFraction(avgFPSStr)
	if err != nil || avgFPS <= 0 {
		return nil, nil, errors.New("Use a constant-frame-rate clean montage before packaging")
	}
	rFPS, err := parseFraction(rFPSStr)
	if err == nil && math.Abs(avgFPS-rFPS) > 0.01 {
		return nil, nil, errors.New("Use a constant-frame-rate clean montage before packaging")
	}

	w := int(videoStream["width"].(float64))
	h := int(videoStream["height"].(float64))
	if w%2 != 0 || h%2 != 0 {
		return nil, nil, errors.New("H.264 delivery requires even dimensions")
	}

	getDuration := func(s, f map[string]any) float64 {
		if durStr, ok := s["duration"].(string); ok {
			if d, err := strconv.ParseFloat(durStr, 64); err == nil && d > 0 {
				return d
			}
		}
		if f != nil {
			if durStr, ok := f["duration"].(string); ok {
				if d, err := strconv.ParseFloat(durStr, 64); err == nil && d > 0 {
					return d
				}
			}
		}
		return 0
	}

	vFormat, _ := vp["format"].(map[string]any)
	aFormat, _ := ap["format"].(map[string]any)
	vDur := getDuration(videoStream, vFormat)
	aDur := getDuration(soundStream, aFormat)
	if math.Abs(aDur-vDur) > 0.4 {
		return nil, nil, errors.New("Picture and approved mix differ by more than 0.4 seconds; synchronize first")
	}

	scriptTokens := Tokens(cfg.Script)
	if len(cfg.Cues) == 0 || len(cfg.Words) == 0 || len(scriptTokens) == 0 {
		return nil, nil, errors.New("Nonempty approved script, semantic cues and word alignment required")
	}

	if err := validateSequence(cfg.Cues, func(c CaptionCue) (float64, float64) { return c.Start, c.End }, vDur, "cues"); err != nil {
		return nil, nil, err
	}
	if err := validateSequence(cfg.Words, func(w WordCue) (float64, float64) { return w.Start, w.End }, vDur, "words"); err != nil {
		return nil, nil, err
	}

	var cueTokens []string
	for _, c := range cfg.Cues {
		cueTokens = append(cueTokens, Tokens(c.Text)...)
	}
	var alignedTokens []string
	for _, w := range cfg.Words {
		wt := Tokens(w.Word)
		if len(wt) != 1 {
			return nil, nil, errors.New("Each word-alignment entry must contain exactly one token")
		}
		alignedTokens = append(alignedTokens, wt...)
	}

	if !sliceEqual(scriptTokens, cueTokens) || !sliceEqual(scriptTokens, alignedTokens) {
		return nil, nil, errors.New("Script/cue/alignment mismatch; correct ASR and numeral spellings first")
	}

	cursor := 0
	for i := range cfg.Cues {
		length := len(Tokens(cfg.Cues[i].Text))
		if length == 0 {
			return nil, nil, errors.New("Use one or two nonempty semantic lines per cue")
		}
		lines := strings.Split(cfg.Cues[i].Text, "\n")
		if len(lines) > 2 {
			return nil, nil, errors.New("Use one or two nonempty semantic lines per cue")
		}
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				return nil, nil, errors.New("Empty caption line")
			}
		}
		cfg.Cues[i].Words = cfg.Words[cursor : cursor+length]
		cursor += length
		firstWord := cfg.Cues[i].Words[0]
		lastWord := cfg.Cues[i].Words[len(cfg.Cues[i].Words)-1]
		if firstWord.Start < cfg.Cues[i].Start-0.12 || lastWord.End > cfg.Cues[i].End+0.12 {
			return nil, nil, fmt.Errorf("Cue cuts off aligned speech: %s", cfg.Cues[i].Text)
		}
	}

	if _, err := parseTitleLines(cfg.Title); err != nil {
		return nil, nil, err
	}

	prepared := &PreparedMedia{
		Config:      &cfg,
		Picture:     picture,
		Audio:       audio,
		CaptionFont: cfg.CaptionFont,
		TitleFont:   cfg.TitleFont,
		Width:       w,
		Height:      h,
		FPS:         avgFPS,
		FPSStr:      avgFPSStr,
		Duration:    vDur,
		Sound:       soundStream,
	}

	return &cfg, prepared, nil
}

// TitleLayer renders 1–2 line uppercase title with dark stroke and white accent bar on left.
func TitleLayer(title any, w, h int, fontPath, emphasisPath string) (*image.NRGBA, error) {
	lines, err := parseTitleLines(title)
	if err != nil {
		return nil, err
	}

	scale := math.Min(float64(w)/720.0, float64(h)/1280.0)
	layer := image.NewNRGBA(image.Rect(0, 0, w, h))

	x := int(math.Round(float64(w) * 62.0 / 720.0))
	y := int(math.Round(float64(h) * 88.0 / 1280.0))

	strokeColor := color.NRGBA{R: 12, G: 10, B: 8, A: 245}
	fillColor := color.NRGBA{R: 255, G: 249, B: 235, A: 255}

	for index, text := range lines {
		size := int(math.Round(35.0 * scale))
		if index != 0 {
			size = int(math.Round(46.0 * scale))
		}
		if size < 1 {
			size = 1
		}
		text = strings.ToUpper(text)
		lw := measureText(text, size)
		if float64(x)+lw > float64(w)*0.94 {
			return nil, errors.New("Title too wide: rephrase or wrap within two lines before rendering")
		}
		lineY := float64(y) + math.Round(float64(index)*38.0*scale)
		strokeWidth := int(math.Max(1.0, math.Round(2.0*scale)))
		drawString(layer, text, float64(x), lineY, size, fillColor, strokeColor, strokeWidth)
	}

	barX1 := int(math.Round(float64(w) * 45.0 / 720.0))
	barX2 := int(math.Round(float64(w) * 49.0 / 720.0))
	barH := 42.0
	if len(lines) == 2 {
		barH = 94.0
	}
	barY2 := y + int(math.Round(barH*scale))

	for by := y; by < barY2; by++ {
		for bx := barX1; bx < barX2; bx++ {
			layer.SetNRGBA(bx, by, fillColor)
		}
	}

	return layer, nil
}

// CaptionLayer renders 1–2 line uppercase caption at bottom, drawing inactive word tokens in cream and active word in amber.
func CaptionLayer(text string, activeIndex int, w, h int, fontPath string) (*image.NRGBA, error) {
	scale := math.Min(float64(w)/720.0, float64(h)/1280.0)
	height := int(math.Round(float64(h) * 250.0 / 1280.0))
	lines := strings.Split(strings.ToUpper(text), "\n")

	size := int(math.Round(76.0 * scale))
	for {
		maxW := 0.0
		for _, line := range lines {
			lw := measureText(line, size)
			if lw > maxW {
				maxW = lw
			}
		}
		if maxW <= float64(w)*618.0/720.0 {
			break
		}
		size--
		if size < int(math.Max(1.0, math.Round(52.0*scale))) {
			return nil, errors.New("Caption too wide; choose a better semantic break, not tiny text")
		}
	}

	layer := image.NewNRGBA(image.Rect(0, 0, w, height))
	lineheight := float64(size) + math.Round(12.0*scale)
	firstY := (float64(height) - float64(len(lines))*lineheight) / 2.0

	strokeWidth := int(math.Max(1.0, math.Round(5.0*scale)))
	strokeColor := color.NRGBA{R: 12, G: 10, B: 8, A: 245}
	inactiveFill := color.NRGBA{R: 255, G: 249, B: 235, A: 255}
	activeFill := color.NRGBA{R: 255, G: 203, B: 74, A: 255}

	tokenIndex := 0
	for j, line := range lines {
		lw := measureText(line, size)
		lineX := (float64(w) - lw) / 2.0
		lineY := firstY + float64(j)*lineheight

		drawString(layer, line, lineX, lineY, size, inactiveFill, strokeColor, strokeWidth)

		matches := tokenRegex.FindAllStringIndex(line, -1)
		for _, m := range matches {
			if tokenIndex == activeIndex {
				prefix := line[:m[0]]
				tokenWord := line[m[0]:m[1]]
				tokenX := lineX + measureText(prefix, size)
				drawString(layer, tokenWord, tokenX, lineY, size, activeFill, strokeColor, strokeWidth)
			}
			tokenIndex++
		}
	}

	return layer, nil
}

func resizeImage(src *image.NRGBA, targetW, targetH int) *image.NRGBA {
	if targetW <= 0 || targetH <= 0 {
		return src
	}
	dst := image.NewNRGBA(image.Rect(0, 0, targetW, targetH))
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	for dy := 0; dy < targetH; dy++ {
		sy := float64(dy) * float64(srcH) / float64(targetH)
		y0 := int(sy)
		y1 := y0 + 1
		if y1 >= srcH {
			y1 = srcH - 1
		}
		fy := sy - float64(y0)
		for dx := 0; dx < targetW; dx++ {
			sx := float64(dx) * float64(srcW) / float64(targetW)
			x0 := int(sx)
			x1 := x0 + 1
			if x1 >= srcW {
				x1 = srcW - 1
			}
			fx := sx - float64(x0)

			c00 := src.NRGBAAt(x0, y0)
			c10 := src.NRGBAAt(x1, y0)
			c01 := src.NRGBAAt(x0, y1)
			c11 := src.NRGBAAt(x1, y1)

			r := (1-fx)*(1-fy)*float64(c00.R) + fx*(1-fy)*float64(c10.R) + (1-fx)*fy*float64(c01.R) + fx*fy*float64(c11.R)
			g := (1-fx)*(1-fy)*float64(c00.G) + fx*(1-fy)*float64(c10.G) + (1-fx)*fy*float64(c01.G) + fx*fy*float64(c11.G)
			b := (1-fx)*(1-fy)*float64(c00.B) + fx*(1-fy)*float64(c10.B) + (1-fx)*fy*float64(c01.B) + fx*fy*float64(c11.B)
			a := (1-fx)*(1-fy)*float64(c00.A) + fx*(1-fy)*float64(c10.A) + (1-fx)*fy*float64(c01.A) + fx*fy*float64(c11.A)

			dst.SetNRGBA(dx, dy, color.NRGBA{
				R: uint8(math.Round(r)),
				G: uint8(math.Round(g)),
				B: uint8(math.Round(b)),
				A: uint8(math.Round(a)),
			})
		}
	}
	return dst
}

func compositeRGB(frame []byte, w, h int, overlay *image.NRGBA, offsetX, offsetY int) {
	bounds := overlay.Bounds()
	ow := bounds.Dx()
	oh := bounds.Dy()
	for oy := 0; oy < oh; oy++ {
		fy := offsetY + oy
		if fy < 0 || fy >= h {
			continue
		}
		for ox := 0; ox < ow; ox++ {
			fx := offsetX + ox
			if fx < 0 || fx >= w {
				continue
			}
			srcColor := overlay.NRGBAAt(ox, oy)
			if srcColor.A == 0 {
				continue
			}
			idx := (fy*w + fx) * 3
			if srcColor.A == 255 {
				frame[idx] = srcColor.R
				frame[idx+1] = srcColor.G
				frame[idx+2] = srcColor.B
			} else {
				alpha := float64(srcColor.A) / 255.0
				invAlpha := 1.0 - alpha
				frame[idx] = uint8(math.Round(float64(srcColor.R)*alpha + float64(frame[idx])*invAlpha))
				frame[idx+1] = uint8(math.Round(float64(srcColor.G)*alpha + float64(frame[idx+1])*invAlpha))
				frame[idx+2] = uint8(math.Round(float64(srcColor.B)*alpha + float64(frame[idx+2])*invAlpha))
			}
		}
	}
}

func frameToRGBA(frame []byte, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sIdx := (y*w + x) * 3
			dIdx := (y*w + x) * 4
			img.Pix[dIdx] = frame[sIdx]
			img.Pix[dIdx+1] = frame[sIdx+1]
			img.Pix[dIdx+2] = frame[sIdx+2]
			img.Pix[dIdx+3] = 255
		}
	}
	return img
}

func saveJPEG(path string, img image.Image, quality int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
}

// Render packages clean picture and approved audio into an A6 deliverable MP4, proofs, captions, and manifest.
func Render(configPath string, outputDir string, validateOnly bool) (map[string]any, error) {
	cfg, media, err := Prepare(configPath)
	if err != nil {
		return nil, err
	}

	title, err := TitleLayer(cfg.Title, media.Width, media.Height, media.TitleFont, media.CaptionFont)
	if err != nil {
		return nil, err
	}

	for _, cue := range cfg.Cues {
		if _, err := CaptionLayer(cue.Text, -1, media.Width, media.Height, media.CaptionFont); err != nil {
			return nil, err
		}
	}

	if validateOnly {
		return map[string]any{
			"validated": true,
			"width":     media.Width,
			"height":    media.Height,
			"fps":       media.FPSStr,
			"duration":  media.Duration,
		}, nil
	}

	if outputDir == "" {
		return nil, errors.New("--output-dir required unless --validate-only")
	}

	out, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, err
	}

	if fi, err := os.Stat(out); err == nil && fi.IsDir() {
		entries, err := os.ReadDir(out)
		if err == nil && len(entries) > 0 {
			return nil, errors.New("Output directory must be new or empty; existing work is preserved")
		}
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return nil, err
	}

	ffmpeg := os.Getenv("FFMPEG")
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}

	totalFrames := int(math.Round(media.Duration * media.FPS))
	proofIndices := make(map[int]bool)
	for _, f := range []float64{0.04, 0.23, 0.57, 0.87} {
		idx := int(math.Round(float64(totalFrames) * f))
		if idx >= totalFrames {
			idx = totalFrames - 1
		}
		if idx >= 0 {
			proofIndices[idx] = true
		}
	}

	captionedMP4 := filepath.Join(out, "picture-captioned.mp4")
	renderLogPath := filepath.Join(out, "render.log")
	renderLog, err := os.OpenFile(renderLogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	defer renderLog.Close()

	decoder := exec.Command(ffmpeg, "-v", "error", "-i", media.Picture, "-an", "-f", "rawvideo", "-pix_fmt", "rgb24", "-")
	decoder.Stderr = renderLog
	decOut, err := decoder.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := decoder.Start(); err != nil {
		return nil, err
	}
	defer func() {
		if decoder.Process != nil {
			_ = decoder.Process.Kill()
		}
	}()

	encoder := exec.Command(ffmpeg, "-v", "warning", "-f", "rawvideo", "-pix_fmt", "rgb24",
		"-s", fmt.Sprintf("%dx%d", media.Width, media.Height), "-r", media.FPSStr, "-i", "-",
		"-an", "-c:v", "libx264", "-crf", "18", "-preset", "fast", "-pix_fmt", "yuv420p", captionedMP4)
	encoder.Stderr = renderLog
	encIn, err := encoder.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := encoder.Start(); err != nil {
		return nil, err
	}
	defer func() {
		if encoder.Process != nil {
			_ = encoder.Process.Kill()
		}
	}()

	rawFrameSize := media.Width * media.Height * 3
	frameBuf := make([]byte, rawFrameSize)
	captionCache := make(map[int]*image.NRGBA)
	var proofPaths []string
	sampleIndex := 1

	cueIndex := 0
	for n := 0; n < totalFrames; n++ {
		_, err := io.ReadFull(decOut, frameBuf)
		if err != nil {
			return nil, fmt.Errorf("Source ended at frame %d, expected %d", n, totalFrames)
		}

		compositeRGB(frameBuf, media.Width, media.Height, title, 0, 0)

		for cueIndex < len(cfg.Cues) && n >= int(math.Round(cfg.Cues[cueIndex].End*media.FPS)) {
			cueIndex++
			captionCache = make(map[int]*image.NRGBA)
		}

		if cueIndex < len(cfg.Cues) {
			cue := cfg.Cues[cueIndex]
			if n >= int(math.Round(cue.Start*media.FPS)) {
				t := float64(n) / media.FPS
				active := -1
				for idx, w := range cue.Words {
					if w.Start <= t {
						active = idx
					}
				}

				layer, ok := captionCache[active]
				if !ok {
					var err error
					layer, err = CaptionLayer(cue.Text, active, media.Width, media.Height, media.CaptionFont)
					if err != nil {
						return nil, err
					}
					captionCache[active] = layer
				}

				age := (t - cue.Start) * 24.0
				pop := 1.0
				if age < 3.0 {
					popStages := []float64{0.92, 1.04, 1.01}
					idx := int(age)
					if idx < 0 {
						idx = 0
					}
					if idx > 2 {
						idx = 2
					}
					pop = popStages[idx]
				}

				var drawLayer *image.NRGBA
				if pop != 1.0 {
					targetW := int(math.Round(float64(media.Width) * pop))
					targetH := int(math.Round(float64(layer.Bounds().Dy()) * pop))
					drawLayer = resizeImage(layer, targetW, targetH)
				} else {
					drawLayer = layer
				}

				layerH := drawLayer.Bounds().Dy()
				layerW := drawLayer.Bounds().Dx()
				posY := int(math.Round(float64(media.Height)*965.0/1280.0 + (float64(media.Height)*250.0/1280.0-float64(layerH))/2.0))
				posX := (media.Width - layerW) / 2
				compositeRGB(frameBuf, media.Width, media.Height, drawLayer, posX, posY)
			}
		}

		if proofIndices[n] {
			proofPath := filepath.Join(out, fmt.Sprintf("proof-%06d.jpg", n))
			samplePath := filepath.Join(out, fmt.Sprintf("sample-%02d.jpg", sampleIndex))
			sampleIndex++
			frameImg := frameToRGBA(frameBuf, media.Width, media.Height)
			if err := saveJPEG(proofPath, frameImg, 94); err != nil {
				return nil, err
			}
			_ = saveJPEG(samplePath, frameImg, 94)
			proofPaths = append(proofPaths, proofPath)
		}

		if _, err := encIn.Write(frameBuf); err != nil {
			return nil, err
		}
	}

	_ = encIn.Close()
	if err := decoder.Wait(); err != nil {
		return nil, errors.New("Render failed; inspect render.log")
	}
	if err := encoder.Wait(); err != nil {
		return nil, errors.New("Render failed; inspect render.log")
	}

	finalMP4 := filepath.Join(out, "a6-final.mp4")
	codecName, _ := media.Sound["codec_name"].(string)
	copyAudio := (codecName == "aac")

	muxArgs := []string{"-v", "warning", "-i", captionedMP4, "-i", media.Audio, "-map", "0:v:0", "-map", "1:a:0", "-c:v", "copy"}
	if copyAudio {
		muxArgs = append(muxArgs, "-c:a", "copy")
	} else {
		muxArgs = append(muxArgs, "-c:a", "aac", "-b:a", "192k")
	}
	muxArgs = append(muxArgs, "-movflags", "+faststart", finalMP4)

	muxCmd := exec.Command(ffmpeg, muxArgs...)
	muxCmd.Stderr = renderLog
	if err := muxCmd.Run(); err != nil {
		return nil, fmt.Errorf("Final muxing failed: %w", err)
	}

	checkCmd := exec.Command(ffmpeg, "-v", "error", "-i", finalMP4, "-f", "null", "-")
	if err := checkCmd.Run(); err != nil {
		return nil, fmt.Errorf("Final MP4 decode check failed: %w", err)
	}

	var srtEntries []string
	for i, c := range cfg.Cues {
		srtEntries = append(srtEntries, fmt.Sprintf("%d\n%s --> %s\n%s", i+1, Stamp(c.Start), Stamp(c.End), strings.ToUpper(c.Text)))
	}
	srtContent := strings.Join(srtEntries, "\n\n") + "\n"
	if err := os.WriteFile(filepath.Join(out, "captions.srt"), []byte(srtContent), 0644); err != nil {
		return nil, err
	}

	sheet := image.NewRGBA(image.Rect(0, 0, 720, 1280))
	for y := 0; y < 1280; y++ {
		for x := 0; x < 720; x++ {
			sheet.Set(x, y, color.RGBA{R: 0x15, G: 0x11, B: 0x17, A: 255})
		}
	}
	for i, pPath := range proofPaths {
		pf, err := os.Open(pPath)
		if err == nil {
			tileImg, err := jpeg.Decode(pf)
			pf.Close()
			if err == nil {
				tileX := (i % 2) * 360
				tileY := (i / 2) * 640
				tw := tileImg.Bounds().Dx()
				th := tileImg.Bounds().Dy()
				for cy := 0; cy < 640; cy++ {
					sy := int(float64(cy) * float64(th) / 640.0)
					for cx := 0; cx < 360; cx++ {
						sx := int(float64(cx) * float64(tw) / 360.0)
						sheet.Set(tileX+cx, tileY+cy, tileImg.At(sx, sy))
					}
				}
			}
		}
	}
	_ = saveJPEG(filepath.Join(out, "contact.jpg"), sheet, 93)

	finalAudioHash, err := AudioHash(finalMP4)
	if err != nil {
		return nil, err
	}
	if copyAudio {
		origAudioHash, err := AudioHash(media.Audio)
		if err != nil {
			return nil, err
		}
		if finalAudioHash != origAudioHash {
			return nil, errors.New("Approved AAC packet preservation check failed")
		}
	}

	finalSha, err := Sha(finalMP4)
	if err != nil {
		return nil, err
	}

	finalProbe, err := Probe(finalMP4)
	if err != nil {
		return nil, err
	}

	sourceSha := func(p string) map[string]string {
		s, _ := Sha(p)
		return map[string]string{"path": p, "sha256": s}
	}

	absCfg, _ := filepath.Abs(configPath)
	manifest := map[string]any{
		"updated_at":           time.Now().UTC().Format(time.RFC3339),
		"status":               "technical_pass_pending_visual_listening_review",
		"final":                finalMP4,
		"sha256":               finalSha,
		"decode_pass":          true,
		"audio_packets_copied": copyAudio,
		"audio_adts_sha256":    finalAudioHash,
		"creator_watermark":    false,
		"generation_calls":     0,
		"probe":                finalProbe,
		"sources": map[string]any{
			"config":       sourceSha(absCfg),
			"picture":      sourceSha(media.Picture),
			"audio":        sourceSha(media.Audio),
			"caption_font": sourceSha(media.CaptionFont),
			"title_font":   sourceSha(media.TitleFont),
		},
		"checks_remaining": []string{
			"semantic segmentation",
			"speech alignment/listening",
			"mobile readability/safe zones",
			"source motion/continuity",
		},
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(out, "manifest.json"), manifestBytes, 0644); err != nil {
		return nil, err
	}

	return map[string]any{
		"final":                finalMP4,
		"sha256":               finalSha,
		"decode_pass":          true,
		"audio_packets_copied": copyAudio,
	}, nil
}
