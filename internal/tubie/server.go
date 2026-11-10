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
	mux.HandleFunc(s.prefix+"api/ai-generate", s.handleAIGenerate)
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
	defer s.mu.RUnlock()

	// List files in output directory
	outDir := filepath.Join(s.root, "output")
	outputs := make([]map[string]any, 0)
	if entries, err := os.ReadDir(outDir); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".mp4") && !strings.HasPrefix(entry.Name(), ".") {
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
				srtPath := filepath.Join(outDir, base+".srt")
				if _, err := os.Stat(srtPath); err == nil {
					item["subtitle"] = "output/" + base + ".srt"
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

	resp := map[string]any{
		"job":       s.job,
		"logs":      s.logs,
		"outputs":   outputs,
		"reference": hasRef,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *StudioServer) handlePresets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(PresetList())
}

func (s *StudioServer) handleMedia(w http.ResponseWriter, r *http.Request) {
	mediaRel := strings.TrimPrefix(r.URL.Path, s.prefix+"media/")
	if strings.Contains(mediaRel, "..") || strings.Contains(mediaRel, "\\") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	parts := strings.Split(mediaRel, "/")
	allowed := false

	if len(parts) == 2 && parts[0] == "output" {
		ext := filepath.Ext(parts[1])
		if ext == ".mp4" || ext == ".srt" || ext == ".png" || ext == ".wav" {
			allowed = true
		}
	} else if mediaRel == "sesim.wav" {
		allowed = true
	} else if len(parts) == 4 && parts[0] == "projects" && parts[2] == "output" {
		if parts[3] == "preview.mp4" || parts[3] == "live-frame.jpg" {
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Prompt string `json:"prompt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Geçersiz JSON verisi"}`, http.StatusBadRequest)
		return
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		http.Error(w, `{"error":"Lütfen bir konu yazın"}`, http.StatusBadRequest)
		return
	}

	s.log(fmt.Sprintf("Yapay zeka senaryo üretiyor: %q...", prompt))
	result, err := GenerateScript(prompt)
	if err != nil {
		s.log(fmt.Sprintf("Yapay zeka hatası: %v", err))
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	s.log(fmt.Sprintf("Senaryo hazır: %s (%d sahne)", result.Title, len(result.Scenes)))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"result": result,
	})
}

func (s *StudioServer) handleJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Action  string   `json:"action"`
		Presets []string `json:"presets"`
		Mode    string   `json:"mode"`
		Custom  *struct {
			Title     string            `json:"title"`
			Slug      string            `json:"slug"`
			Narration string            `json:"narration"`
			Scenes    []SceneDefinition `json:"scenes"`
		} `json:"custom,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Geçersiz JSON verisi"}`, http.StatusBadRequest)
		return
	}

	var customPreset *Preset
	if req.Custom != nil && strings.TrimSpace(req.Custom.Narration) != "" {
		slug := req.Custom.Slug
		if slug == "" {
			slug = slugify(req.Custom.Title)
		}
		customPreset = &Preset{
			ID:        "custom",
			Title:     req.Custom.Title,
			Slug:      slug,
			Narration: req.Custom.Narration,
			Scenes:    req.Custom.Scenes,
		}
	}

	s.mu.Lock()
	if s.job["status"] == "running" || s.job["status"] == "approval" {
		s.mu.Unlock()
		http.Error(w, `{"error":"Devam eden işlem tamamlanmalı veya iptal edilmeli"}`, http.StatusBadRequest)
		return
	}

	s.cancelCtx, s.cancelFunc = context.WithCancel(context.Background())
	s.job = map[string]any{
		"status":  "running",
		"stage":   "starting",
		"label":   "İşlem başlatılıyor",
		"action":  req.Action,
		"results": []string{},
	}
	s.mu.Unlock()

	go s.runJob(req.Action, req.Presets, req.Mode, customPreset)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *StudioServer) runJob(action string, presets []string, mode string, customPreset *Preset) {
	ctx := s.cancelCtx

	defer func() {
		s.mu.Lock()
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

	if action == "generate" {
		pipeline := &Pipeline{
			Root: s.root,
			Mode: mode,
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

			s.log(fmt.Sprintf("[%d/%d] Üretim başlatılıyor: %s", idx+1, len(jobsToRun), preset.Title))
			out, err := pipeline.Produce(preset)
			if err != nil {
				s.log(fmt.Sprintf("HATA: %v", err))
				s.mu.Lock()
				s.job["status"] = "failed"
				s.job["label"] = fmt.Sprintf("Hata: %v", err)
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
	select {
	case s.approveChan <- struct{}{}:
	default:
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *StudioServer) handleCancel(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.cancelFunc != nil {
		s.job["status"] = "stopping"
		s.job["label"] = "İşlem durduruluyor"
		s.cancelFunc()
	}
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *StudioServer) handleVoiceUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 50*1024*1024))
	if err != nil || len(body) == 0 {
		http.Error(w, `{"error":"Boş veya geçersiz ses verisi"}`, http.StatusBadRequest)
		return
	}

	tempFile, err := os.CreateTemp("", "tubie-upload-*.audio")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Temp file error: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(body); err != nil {
		tempFile.Close()
		http.Error(w, `{"error":"Yazma hatası"}`, http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	dest := filepath.Join(s.root, "sesim.wav")
	if err := ConvertReference(tempFile.Name(), dest); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Ses dönüştürme hatası: %v"}`, err), http.StatusBadRequest)
		return
	}

	s.log("Yeni ses referansı yüklendi: sesim.wav")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *StudioServer) handleShutdown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))

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

