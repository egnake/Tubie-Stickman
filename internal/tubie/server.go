package tubie

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

//go:embed web/*
var embeddedWeb embed.FS

// StudioServer manages the local-only Tubie Studio HTTP interface.
type StudioServer struct {
	root     string
	token    string
	prefix   string
	listener net.Listener
	server   *http.Server
	port     int

	mu       sync.RWMutex
	job      map[string]any
	logs     []map[string]any
	revision int

	approveChan chan struct{}
	cancelCtx   context.Context
	cancelFunc  context.CancelFunc
}

// NewStudioServer initializes a new server instance.
func NewStudioServer(root string, port int) (*StudioServer, error) {
	tokenBytes := make([]byte, 18)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port

	s := &StudioServer{
		root:        root,
		token:       token,
		prefix:      "/" + token + "/",
		listener:    listener,
		port:        actualPort,
		approveChan: make(chan struct{}, 1),
		logs:        make([]map[string]any, 0),
		job: map[string]any{
			"status":  "idle",
			"stage":   "idle",
			"label":   "Yeni bir üretime hazır",
			"results": []string{},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(s.prefix, s.handleRoot)
	mux.HandleFunc(s.prefix+"api/state", s.handleState)
	mux.HandleFunc(s.prefix+"api/presets", s.handlePresets)
	mux.HandleFunc(s.prefix+"api/ideas", s.handleIdeas)
	mux.HandleFunc(s.prefix+"api/ai-generate", s.handleAIGenerate)
	mux.HandleFunc(s.prefix+"api/viral-analyze", s.handleViralAnalyze)
	mux.HandleFunc(s.prefix+"api/thumbnail-preview", s.handleThumbnailPreview)
	mux.HandleFunc(s.prefix+"api/job", s.handleJob)
	mux.HandleFunc(s.prefix+"api/approve", s.handleApprove)
	mux.HandleFunc(s.prefix+"api/cancel", s.handleCancel)
	mux.HandleFunc(s.prefix+"api/voice", s.handleVoiceUpload)
	mux.HandleFunc(s.prefix+"api/shutdown", s.handleShutdown)
	mux.HandleFunc(s.prefix+"media/", s.handleMedia)

	s.server = &http.Server{
		Handler: s.securityMiddleware(mux),
	}

	return s, nil
}

// URL returns the capability URL for this studio instance.
func (s *StudioServer) URL(tab string) string {
	if tab == "" {
		tab = "produce"
	}
	return fmt.Sprintf("http://127.0.0.1:%d%s#%s", s.port, s.prefix, tab)
}

func (s *StudioServer) log(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revision++
	s.logs = append(s.logs, map[string]any{
		"id":   s.revision,
		"time": time.Now().Format("15:04:05"),
		"text": text,
	})
	if len(s.logs) > 300 {
		s.logs = s.logs[len(s.logs)-300:]
	}
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeLimitedJSON(w http.ResponseWriter, r *http.Request, maxBytes int64, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Geçersiz JSON verisi"})
		return false
	}
	return true
}

func cloneJobMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case []string:
			cp := make([]string, len(val))
			copy(cp, val)
			dst[k] = cp
		default:
			dst[k] = val
		}
	}
	return dst
}

func (s *StudioServer) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := fmt.Sprintf("127.0.0.1:%d", s.port)
		if r.Host != host {
			http.Error(w, "Forbidden Host", http.StatusForbidden)
			return
		}

		origin := r.Header.Get("Origin")
		if origin != "" && origin != "http://"+host {
			http.Error(w, "Forbidden Origin", http.StatusForbidden)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self' blob:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'")

		next.ServeHTTP(w, r)
	})
}

func (s *StudioServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	rel := strings.TrimPrefix(r.URL.Path, s.prefix)
	if rel == "" || rel == "index.html" {
		s.serveWebFile(w, r, "index.html", "text/html; charset=utf-8")
		return
	}
	if rel == "app.css" {
		s.serveWebFile(w, r, "app.css", "text/css; charset=utf-8")
		return
	}
	if rel == "app.js" {
		s.serveWebFile(w, r, "app.js", "application/javascript; charset=utf-8")
		return
	}
	http.NotFound(w, r)
}

func (s *StudioServer) serveWebFile(w http.ResponseWriter, r *http.Request, name string, contentType string) {
	// Try disk first
	diskPath := filepath.Join(s.root, "internal", "tubie", "web", name)
	if data, err := os.ReadFile(diskPath); err == nil {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		return
	}

	// Try embedded
	if data, err := embeddedWeb.ReadFile("web/" + name); err == nil {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		return
	}

	http.NotFound(w, r)
}

func (s *StudioServer) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	jobSnap := cloneJobMap(s.job)
	logsSnap := make([]map[string]any, len(s.logs))
	copy(logsSnap, s.logs)
	s.mu.RUnlock()

	// List files in output directory (both organized subdirectories and legacy root files)
	outDir := filepath.Join(s.root, "output")
	outputs := make([]map[string]any, 0)
	if entries, err := os.ReadDir(outDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				slug := entry.Name()
				subDir := filepath.Join(outDir, slug)
				if subEntries, err := os.ReadDir(subDir); err == nil {
					for _, se := range subEntries {
						if strings.HasSuffix(se.Name(), ".mp4") && !strings.HasPrefix(se.Name(), ".") {
							base := strings.TrimSuffix(se.Name(), ".mp4")
							fi, _ := se.Info()
							size := int64(0)
							if fi != nil {
								size = fi.Size()
							}
							item := map[string]any{
								"name": strings.ReplaceAll(slug, "_", " "),
								"file": fmt.Sprintf("output/%s/%s", slug, se.Name()),
								"size": size,
							}
							if isFile(filepath.Join(subDir, base+".srt")) {
								item["subtitle"] = fmt.Sprintf("output/%s/%s.srt", slug, base)
							}
							if isFile(filepath.Join(subDir, base+".vtt")) {
								item["vtt"] = fmt.Sprintf("output/%s/%s.vtt", slug, base)
							}
							if isFile(filepath.Join(subDir, "thumbnail.jpg")) {
								item["thumbnail"] = fmt.Sprintf("output/%s/thumbnail.jpg", slug)
							}
							if isFile(filepath.Join(subDir, "thumbnail_b.jpg")) {
								item["thumbnail_b"] = fmt.Sprintf("output/%s/thumbnail_b.jpg", slug)
							}
							if isFile(filepath.Join(subDir, "thumbnail_16x9.jpg")) {
								item["thumbnail_16x9"] = fmt.Sprintf("output/%s/thumbnail_16x9.jpg", slug)
							}
							txtPath := filepath.Join(subDir, "aciklama.txt")
							if isFile(txtPath) {
								item["description"] = fmt.Sprintf("output/%s/aciklama.txt", slug)
								if descBytes, err := os.ReadFile(txtPath); err == nil {
									item["description_text"] = string(descBytes)
								}
							}
							reportPath := filepath.Join(subDir, "viral_report.json")
							if isFile(reportPath) {
								if rBytes, err := os.ReadFile(reportPath); err == nil {
									var vr map[string]any
									if err := json.Unmarshal(rBytes, &vr); err == nil {
										item["viral_report"] = vr
										item["viral_score"] = vr["total_score"]
										item["viral_grade"] = vr["grade"]
										item["format"] = vr["format"]
									}
								}
							}
							outputs = append(outputs, item)
						}
					}
				}
			} else if strings.HasSuffix(entry.Name(), ".mp4") && !strings.HasPrefix(entry.Name(), ".") {
				base := strings.TrimSuffix(entry.Name(), ".mp4")
				fi, _ := entry.Info()
				size := int64(0)
				if fi != nil {
					size = fi.Size()
				}
				item := map[string]any{
					"name": strings.ReplaceAll(base, "_", " "),
					"file": "output/" + entry.Name(),
					"size": size,
				}
				if isFile(filepath.Join(outDir, base+".srt")) {
					item["subtitle"] = "output/" + base + ".srt"
				}
				if isFile(filepath.Join(outDir, base+".vtt")) {
					item["vtt"] = "output/" + base + ".vtt"
				}
				outputs = append(outputs, item)
			}
		}
	}

	refPath := filepath.Join(s.root, "sesim.wav")
	hasRef := false
	if fi, err := os.Stat(refPath); err == nil && !fi.IsDir() && fi.Size() > 0 {
		hasRef = true
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"job":       jobSnap,
		"logs":      logsSnap,
		"outputs":   outputs,
		"reference": hasRef,
	})
}

func (s *StudioServer) handlePresets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, PresetList())
}

func (s *StudioServer) handleIdeas(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"ideas": TrendingCyberIdeas(),
	})
}

func (s *StudioServer) handleMedia(w http.ResponseWriter, r *http.Request) {
	mediaRel := strings.TrimPrefix(r.URL.Path, s.prefix+"media/")
	if strings.Contains(mediaRel, "..") || strings.Contains(mediaRel, "\\") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	parts := strings.Split(mediaRel, "/")
	allowed := false

	if len(parts) >= 2 && parts[0] == "output" {
		filename := parts[len(parts)-1]
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".mp4", ".srt", ".vtt", ".png", ".jpg", ".jpeg", ".wav", ".txt", ".json":
			allowed = true
		}
	} else if mediaRel == "sesim.wav" {
		allowed = true
	} else if len(parts) == 4 && parts[0] == "projects" && parts[2] == "output" {
		switch parts[3] {
		case "preview.mp4", "live-frame.jpg", "preview-9x16.jpg", "preview-16x9.jpg":
			allowed = true
		}
	}

	if !allowed {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	target := filepath.Join(s.root, filepath.FromSlash(mediaRel))
	relFromRoot, err := filepath.Rel(s.root, target)
	if err != nil || strings.HasPrefix(relFromRoot, "..") {
		http.Error(w, "Path traversal", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, target)
}

func (s *StudioServer) handleAIGenerate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		Prompt     string `json:"prompt"`
		Style      string `json:"style"`
		LoopEnding bool   `json:"loop_ending"`
	}
	if !decodeLimitedJSON(w, r, 2<<20, &req) {
		return
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Lütfen bir konu yazın"})
		return
	}

	style := NormalizeScriptStyle(req.Style)
	s.log(fmt.Sprintf("Yapay zeka senaryo üretiyor [%s, loop=%v]: %q...", style, req.LoopEnding, prompt))
	result, err := GenerateScriptWithOptions(prompt, ScriptOptions{
		Style:      style,
		LoopEnding: req.LoopEnding,
	})
	if err != nil {
		s.log(fmt.Sprintf("Yapay zeka hatası: %v", err))
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	s.log(fmt.Sprintf("Senaryo hazır (%s · %s): %s (%d sahne)", result.Style, result.Engine, result.Title, len(result.Scenes)))
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"result": result,
	})
}

func (s *StudioServer) handleViralAnalyze(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		Title      string            `json:"title"`
		Narration  string            `json:"narration"`
		Scenes     []SceneDefinition `json:"scenes"`
		LoopEnding bool              `json:"loop_ending"`
	}
	if !decodeLimitedJSON(w, r, 2<<20, &req) {
		return
	}

	title := strings.TrimSpace(req.Title)
	narration := strings.TrimSpace(req.Narration)
	if title == "" && narration == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Analiz için başlık veya metin gerekli"})
		return
	}

	scenes := req.Scenes
	if len(scenes) == 0 {
		scenes = EnsureScenesForScript(title, narration)
	}
	report := AnalyzeScriptVirality(title, narration, scenes, req.LoopEnding)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"report": report,
	})
}

func (s *StudioServer) handleThumbnailPreview(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		Title string `json:"title"`
		Topic string `json:"topic"`
		Mood  string `json:"mood"`
	}
	if !decodeLimitedJSON(w, r, 2<<20, &req) {
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "SİBER GÜVENLİK UYARISI"
	}
	topic := strings.TrimSpace(req.Topic)
	if topic == "" {
		topic = title
	}
	mood := strings.TrimSpace(req.Mood)
	if mood == "" {
		mood = "danger"
	}

	previewDir := filepath.Join(s.root, "projects", "_thumb_preview", "output")
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	out9x16 := filepath.Join(previewDir, "preview-9x16.jpg")
	if err := GenerateDedicatedThumbnailWithOptions(s.root, title, topic, "", 2, ThumbnailOptions{
		Aspect:  "9:16",
		Variant: "a",
		Mood:    mood,
	}, out9x16, ""); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("9:16 kapak hatası: %v", err)})
		return
	}

	out16x9 := filepath.Join(previewDir, "preview-16x9.jpg")
	if err := GenerateDedicatedThumbnailWithOptions(s.root, title, topic, "", 2, ThumbnailOptions{
		Aspect:  "16:9",
		Variant: "a",
		Mood:    mood,
	}, out16x9, ""); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("16:9 kapak hatası: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"preview_9x16": "projects/_thumb_preview/output/preview-9x16.jpg",
		"preview_16x9": "projects/_thumb_preview/output/preview-16x9.jpg",
		"revision":     time.Now().UnixNano(),
	})
}

func (s *StudioServer) handleJob(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		Action     string   `json:"action"`
		Presets    []string `json:"presets"`
		Mode       string   `json:"mode"`
		VoiceStyle string   `json:"voice_style"`
		BGM        string   `json:"bgm"`
		Format     string   `json:"format"`
		LoopEnding bool     `json:"loop_ending"`
		Custom     *struct {
			Title     string            `json:"title"`
			Slug      string            `json:"slug"`
			Narration string            `json:"narration"`
			Scenes    []SceneDefinition `json:"scenes"`
		} `json:"custom,omitempty"`
	}
	if !decodeLimitedJSON(w, r, 2<<20, &req) {
		return
	}

	var customPreset *Preset
	if req.Custom != nil && strings.TrimSpace(req.Custom.Narration) != "" {
		slug := sanitizeSlug(req.Custom.Slug, req.Custom.Title)
		scenes := req.Custom.Scenes
		if len(scenes) == 0 {
			scenes = EnsureScenesForScript(req.Custom.Title, req.Custom.Narration)
		}
		customPreset = &Preset{
			ID:        "custom",
			Title:     req.Custom.Title,
			Slug:      slug,
			Narration: req.Custom.Narration,
			Scenes:    scenes,
		}
	}

	s.mu.Lock()
	if s.job["status"] == "running" || s.job["status"] == "approval" {
		s.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Devam eden işlem tamamlanmalı veya iptal edilmeli"})
		return
	}

	// Drain stale approval signals from any earlier cancelled/finished run.
	select {
	case <-s.approveChan:
	default:
	}

	fmtSpec := ResolveVideoFormat(req.Format)
	s.cancelCtx, s.cancelFunc = context.WithCancel(context.Background())
	s.job = map[string]any{
		"status":      "running",
		"stage":       "starting",
		"label":       "İşlem başlatılıyor",
		"action":      req.Action,
		"voice_style": req.VoiceStyle,
		"format":      fmtSpec.Name,
		"loop_ending": req.LoopEnding,
		"started":     time.Now().Unix(),
		"results":     []string{},
	}
	s.mu.Unlock()

	go s.runJob(req.Action, req.Presets, req.Mode, req.VoiceStyle, req.BGM, fmtSpec.Name, req.LoopEnding, customPreset)

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *StudioServer) runJob(action string, presets []string, mode, voiceStyle, bgm, format string, loopEnding bool, customPreset *Preset) {
	ctx := s.cancelCtx

	defer func() {
		s.mu.Lock()
		s.job["ended"] = time.Now().Unix()
		if s.job["status"] == "stopping" {
			s.job["status"] = "cancelled"
			s.job["label"] = "İşlem iptal edildi"
		} else if s.job["status"] == "running" {
			s.job["status"] = "complete"
			s.job["label"] = "İşlem tamamlandı"
		}
		s.mu.Unlock()
	}()

	if action == "doctor" {
		s.log("Ortam kontrolü başlatılıyor...")
		code := Doctor(s.root)
		if code == 0 {
			s.log("Ortam kontrolü başarıyla tamamlandı.")
		} else {
			s.log("Bazı bağımlılıklar eksik.")
		}
		return
	}

	if action == "test-voice" {
		if !strings.HasPrefix(voiceStyle, "cloned") {
			voiceStyle = "cloned"
		}
		outDir := filepath.Join(s.root, "output")
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			s.log(fmt.Sprintf("Test sesi klasörü oluşturulamadı: %v", err))
			s.mu.Lock()
			s.job["status"] = "failed"
			s.job["error"] = err.Error()
			s.job["label"] = "Test sesi hatası"
			s.mu.Unlock()
			return
		}
		outWav := filepath.Join(outDir, "test-voice.wav")
		sampleText := "Merhaba! Klonlanmış sesin artık çok daha hızlı, enerjik ve akıcı."
		s.log(fmt.Sprintf("Klonlanmış ses testi başlatılıyor (%s)...", voiceStyle))
		dur, used, err := SynthesizeVoiceContext(ctx, sampleText, outWav, "cloned", voiceStyle, s.root)
		if err != nil {
			s.log(fmt.Sprintf("Ses testi hatası: %v", err))
			s.mu.Lock()
			s.job["status"] = "failed"
			s.job["error"] = err.Error()
			s.job["label"] = "Ses testi başarısız"
			s.mu.Unlock()
			return
		}
		s.mu.Lock()
		s.job["audio"] = "output/test-voice.wav"
		s.mu.Unlock()
		s.log(fmt.Sprintf("Test sesi hazır (%s · %.2f sn): output/test-voice.wav", used, dur))
		return
	}

	if action == "generate" {
		pipeline := &Pipeline{
			Root:       s.root,
			Mode:       mode,
			VoiceStyle: voiceStyle,
			BGM:        bgm,
			Format:     format,
			LoopEnding: loopEnding,
			Reporter: func(ev map[string]any) {
				s.mu.Lock()
				defer s.mu.Unlock()
				for k, v := range ev {
					s.job[k] = v
				}
				if stage, ok := ev["stage"].(string); ok {
					s.job["stage"] = stage
				}
				if label, ok := ev["label"].(string); ok {
					s.job["label"] = label
				}
				if frame, ok := ev["frame"]; ok {
					s.job["frame"] = frame
					if slug, ok := s.job["slug"].(string); ok {
						s.job["live"] = fmt.Sprintf("projects/%s/output/live-frame.jpg", slug)
						s.job["live_revision"] = time.Now().UnixNano()
					}
				}
			},
			Review: func(projectDir string) error {
				slug := filepath.Base(projectDir)
				s.mu.Lock()
				s.job["status"] = "approval"
				s.job["stage"] = "approval"
				s.job["label"] = "Önizleme hazır · onayınızı bekliyor"
				s.job["preview"] = fmt.Sprintf("projects/%s/output/preview.mp4", slug)
				s.mu.Unlock()

				s.log(fmt.Sprintf("Önizleme hazır: %s/output/preview.mp4", slug))

				select {
				case <-s.approveChan:
					s.mu.Lock()
					s.job["status"] = "running"
					s.job["stage"] = "render"
					s.job["label"] = "Son video işleniyor"
					s.mu.Unlock()
					return nil
				case <-ctx.Done():
					return fmt.Errorf("cancelled")
				}
			},
		}

		var jobsToRun []Preset
		if customPreset != nil {
			jobsToRun = append(jobsToRun, *customPreset)
		} else {
			for _, pKey := range presets {
				if preset, ok := CyberPresets[pKey]; ok {
					jobsToRun = append(jobsToRun, preset)
				}
			}
		}

		for idx, preset := range jobsToRun {
			select {
			case <-ctx.Done():
				return
			default:
			}

			s.mu.Lock()
			s.job["item"] = idx + 1
			s.job["items"] = len(jobsToRun)
			s.job["title"] = preset.Title
			s.job["slug"] = preset.Slug
			s.mu.Unlock()

			s.log(fmt.Sprintf("[%d/%d] Üretim başlatılıyor: %s [%s]", idx+1, len(jobsToRun), preset.Title, format))
			out, err := pipeline.ProduceContext(ctx, preset)
			if err != nil {
				s.log(fmt.Sprintf("HATA: %v", err))
				s.mu.Lock()
				if ctx.Err() != nil {
					s.job["status"] = "cancelled"
					s.job["label"] = "İşlem iptal edildi"
				} else {
					s.job["status"] = "failed"
					s.job["error"] = err.Error()
					s.job["label"] = fmt.Sprintf("Hata: %v", err)
				}
				s.mu.Unlock()
				return
			}

			s.mu.Lock()
			rel, _ := filepath.Rel(s.root, out)
			results, _ := s.job["results"].([]string)
			s.job["results"] = append(results, filepath.ToSlash(rel))
			s.mu.Unlock()
			s.log(fmt.Sprintf("Çıktı hazır: %s", out))
		}
	}
}

func (s *StudioServer) handleApprove(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	select {
	case s.approveChan <- struct{}{}:
	default:
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *StudioServer) handleCancel(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	s.mu.Lock()
	if s.cancelFunc != nil {
		s.job["status"] = "stopping"
		s.job["label"] = "İşlem durduruluyor"
		s.cancelFunc()
	}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *StudioServer) handleVoiceUpload(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 50*1024*1024))
	if err != nil || len(body) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Boş veya geçersiz ses verisi"})
		return
	}

	tempFile, err := os.CreateTemp("", "tubie-upload-*.audio")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("Temp file error: %v", err)})
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(body); err != nil {
		tempFile.Close()
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Yazma hatası"})
		return
	}
	tempFile.Close()

	dest := filepath.Join(s.root, "sesim.wav")
	if err := ConvertReference(tempFile.Name(), dest); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("Ses dönüştürme hatası: %v", err)})
		return
	}

	s.log("Yeni ses referansı yüklendi: sesim.wav")
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *StudioServer) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})

	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = s.server.Close()
		os.Exit(0)
	}()
}

// Serve starts listening and serving requests.
func (s *StudioServer) Serve() error {
	return s.server.Serve(s.listener)
}

// Close closes the underlying HTTP server cleanly.
func (s *StudioServer) Close() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// OpenBrowser opens the URL in the system default browser.
func OpenBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
