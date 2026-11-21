package tubie_test

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"tubie-stickman/internal/tubie"
	"unicode/utf8"
)

func TestFindRoot(t *testing.T) {
	root, err := tubie.FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("go.mod not found in detected root: %s", root)
	}
}

func TestPresets(t *testing.T) {
	presets := tubie.PresetList()
	if len(presets) != 10 {
		t.Fatalf("expected 10 presets, got %d", len(presets))
	}
	for _, p := range presets {
		if p.ID == "" || p.Title == "" || p.Slug == "" || p.Narration == "" || p.Category == "" {
			t.Errorf("incomplete preset data: %+v", p)
		}
	}
}

func TestBinaryPath(t *testing.T) {
	nodePath := tubie.BinaryPath("node")
	if nodePath == "" {
		t.Fatal("expected non-empty node path")
	}
}

func TestDoctor(t *testing.T) {
	root, err := tubie.FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}
	code := tubie.Doctor(root)
	if code != 0 {
		t.Fatalf("Doctor returned failure code: %d", code)
	}
}

func TestStudioServerInitialization(t *testing.T) {
	root, err := tubie.FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}
	server, err := tubie.NewStudioServer(root, 0)
	if err != nil {
		t.Fatalf("failed to create StudioServer: %v", err)
	}
	url := server.URL("produce")
	if url == "" || len(url) < 15 {
		t.Fatalf("invalid server URL: %s", url)
	}
}

func TestAccurateCaptions(t *testing.T) {
	script := "Hackerlar şifreni tahmin etmeye çalışmaz. Deauth saldırısıyla seni bir saniyede ağdan düşürürler!"
	captions := tubie.BuildAccurateCaptions(script, "", 10.0)
	if len(captions) == 0 {
		t.Fatal("expected non-empty captions")
	}

	cueEnd := 0.0
	for i, c := range captions {
		if c.Start < cueEnd {
			t.Errorf("cue %d: start %.3f is less than previous end %.3f", i, c.Start, cueEnd)
		}
		if c.End <= c.Start {
			t.Errorf("cue %d: end %.3f <= start %.3f", i, c.End, c.Start)
		}
		if c.End > 10.0 {
			t.Errorf("cue %d: end %.3f exceeds total duration", i, c.End)
		}
		cueEnd = c.End
	}

	srt := tubie.FormatSRT(captions)
	if srt == "" {
		t.Fatal("expected non-empty srt")
	}
	// Verify that hallucinated Whisper text is NOT present
	if strings.Contains(srt, "ETMEYE ÇALIŞMAZ DIT") || strings.Contains(srt, "AKARLAR") {
		t.Fatal("srt contains corrupted Whisper tokens")
	}
	// Verify proper Turkish casing
	if !strings.Contains(srt, "HACKERLAR") || !strings.Contains(srt, "DEAUTH") {
		t.Errorf("srt missing canonical words: %s", srt)
	}
}

func TestBuildStoryboard(t *testing.T) {
	scenes := []tubie.SceneDefinition{
		{Title: "SAHTE DM TUZAĞI", SFX: "keyboard"},
		{Title: "ŞİFREN ELE GEÇİRİLDİ", SFX: "alert"},
		{Title: "HESAP KONTROLÜ KAYBEDİLDİ", SFX: "glitch"},
		{Title: "2FA İLE KORU", SFX: "lock"},
	}
	totalDur := 25.5
	shots := tubie.BuildStoryboard(scenes, totalDur)
	if len(shots) != len(scenes) {
		t.Fatalf("expected %d shots, got %d", len(scenes), len(shots))
	}

	currStart := 0.0
	for i, s := range shots {
		start := s["start"].(float64)
		end := s["end"].(float64)
		span := end - start
		if math.Abs(start-currStart) > 0.002 {
			t.Errorf("shot %d: start %.3f != expected %.3f", i, start, currStart)
		}
		if end <= start {
			t.Errorf("shot %d: end <= start", i)
		}

		// Verify all layers have monotonic keyframe times strictly within [0, span]
		layers := s["layers"].([]map[string]any)
		for li, l := range layers {
			if rawKeys, ok := l["keys"]; ok {
				keys := rawKeys.([]map[string]any)
				prevT := -0.001
				for ki, k := range keys {
					kt := k["t"].(float64)
					if kt <= prevT {
						t.Errorf("shot %d layer %d key %d: non-increasing t=%.6f <= prev=%.6f", i, li, ki, kt, prevT)
					}
					if kt < 0 || kt > span {
						t.Errorf("shot %d layer %d key %d: t=%.6f out of exact shot bounds [0, %.6f]", i, li, ki, kt, span)
					}
					prevT = kt
				}
			}
		}

		currStart = end
	}
	if math.Abs(currStart-totalDur) > 0.002 {
		t.Errorf("final shot end %.3f does not match totalDur %.3f", currStart, totalDur)
	}

	// Verify 6 scenes at float-tricky duration (31.248) and ultra-short duration (1.8)
	sixScenes := []tubie.SceneDefinition{
		{Title: "S1"}, {Title: "S2"}, {Title: "S3"}, {Title: "S4"}, {Title: "S5"}, {Title: "S6"},
	}
	for _, dur := range []float64{31.248, 1.8} {
		checkShots := tubie.BuildStoryboard(sixScenes, dur)
		for i, s := range checkShots {
			span := s["end"].(float64) - s["start"].(float64)
			for li, l := range s["layers"].([]map[string]any) {
				if rawKeys, ok := l["keys"]; ok {
					prevT := -0.001
					for ki, k := range rawKeys.([]map[string]any) {
						kt := k["t"].(float64)
						if kt <= prevT || kt < 0 || kt > span {
							t.Errorf("dur %.3f shot %d layer %d key %d: invalid key t=%.6f (prev=%.6f, span=%.15f)", dur, i, li, ki, kt, prevT, span)
						}
						prevT = kt
					}
				}
			}
		}
	}
}

func TestBuildSoundDesign(t *testing.T) {
	scenes := []tubie.SceneDefinition{
		{Title: "SAHTE DM TUZAĞI", SFX: "keyboard"},
		{Title: "ŞİFREN ELE GEÇİRİLDİ", SFX: "alert"},
		{Title: "HESAP KONTROLÜ KAYBEDİLDİ", SFX: "glitch"},
		{Title: "2FA İLE KORU", SFX: "lock"},
	}
	totalDur := 20.0
	sfx := tubie.BuildSoundDesign(scenes, totalDur)
	if len(sfx) == 0 {
		t.Fatal("expected non-empty sfx cues")
	}

	prevTime := -0.001
	kinds := make(map[string]bool)
	for i, cue := range sfx {
		tVal := cue["time"].(float64)
		gain := cue["gain"].(float64)
		kind := cue["kind"].(string)

		if tVal < prevTime {
			t.Errorf("cue %d: time %.3f is not monotonic (prev %.3f)", i, tVal, prevTime)
		}
		if tVal >= totalDur {
			t.Errorf("cue %d: time %.3f exceeds total duration %.3f", i, tVal, totalDur)
		}
		if gain <= 0 || gain > 0.30 {
			t.Errorf("cue %d: gain %.3f invalid (must be 0..0.30)", i, gain)
		}
		prevTime = tVal
		kinds[kind] = true
	}

	// Verify essential sound design kinds are present
	for _, expectedKind := range []string{"pop", "paper", "jingle"} {
		if !kinds[expectedKind] {
			t.Errorf("expected sound design kind %q was not generated", expectedKind)
		}
	}
}

func TestGenerateScript(t *testing.T) {
	res, err := tubie.GenerateScript("İnstagram hesap çalma")
	if err != nil {
		t.Fatalf("GenerateScript failed: %v", err)
	}
	if res.Title == "" || res.Narration == "" || res.Slug == "" {
		t.Fatalf("incomplete script result: %+v", res)
	}
	if len(res.Scenes) < 3 {
		t.Fatalf("expected at least 3 scenes, got %d", len(res.Scenes))
	}
}

func TestViralMetadata(t *testing.T) {
	title := "İnstagram Hesabınız Nasıl Çalınır?"
	narration := "Instagram hesabın bir gecede elinden gidebilir! Hackerlar şifreni kırmakla uğraşmaz."
	vm := tubie.GenerateViralMetadata(title, narration)

	if vm.TitleA == "" || vm.TitleB == "" || vm.TitleC == "" {
		t.Fatalf("missing A/B/C titles: %+v", vm)
	}
	if vm.Description == "" {
		t.Fatal("missing viral description")
	}
	if len(vm.Hashtags) < 5 {
		t.Fatalf("expected at least 5 hashtags, got %d", len(vm.Hashtags))
	}

	text := tubie.FormatViralText(title, narration, vm)
	if !strings.Contains(text, "[BAŞLIK A") || !strings.Contains(text, "[BAŞLIK B") || !strings.Contains(text, "[BAŞLIK C") {
		t.Fatalf("formatted text missing A/B/C headers: %s", text)
	}
	if !strings.Contains(text, "VİRAL REELS & POST AÇIKLAMASI") {
		t.Fatalf("formatted text missing post description: %s", text)
	}
	if !strings.Contains(text, "VİRAL KEŞFET HASHTAGLERİ") {
		t.Fatalf("formatted text missing hashtags: %s", text)
	}
	if !strings.Contains(text, "YOUTUBE TAG") {
		t.Fatalf("formatted text missing youtube tag header: %s", text)
	}
	if !strings.Contains(text, "atıl samancıoğlu") || !strings.Contains(text, "can değer") || !strings.Contains(text, "mdisec") {
		t.Fatalf("formatted text missing required youtube tag keywords: %s", text)
	}
	// Verify Instagram 2200 character limit on the post caption + hashtags
	postCaption := vm.Description + "\n\n" + strings.Join(vm.Hashtags, " ")
	if len([]rune(postCaption)) > 2200 {
		t.Fatalf("instagram post caption exceeds 2200 character limit: %d characters", len([]rune(postCaption)))
	}
}

func TestGenerateDedicatedThumbnail(t *testing.T) {
	root, err := tubie.FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}

	tmpDir := t.TempDir()
	outThumb := filepath.Join(tmpDir, "test_thumbnail.jpg")

	err = tubie.GenerateDedicatedThumbnail(root, "Instagram Hesabınız Nasıl Çalınır?", "sosyal medya tuzağı", "", 2, outThumb, "")
	if err != nil {
		t.Fatalf("GenerateDedicatedThumbnail failed: %v", err)
	}

	fi, err := os.Stat(outThumb)
	if err != nil {
		t.Fatalf("thumbnail file not found: %v", err)
	}
	if fi.Size() < 5000 {
		t.Fatalf("thumbnail file suspiciously small: %d bytes", fi.Size())
	}
}

func TestClonedVoiceTimingAndFilter(t *testing.T) {
	speedStd, atempoStd := tubie.ClonedVoiceTiming("cloned")
	if speedStd <= 1.12 || atempoStd <= 1.0 {
		t.Fatalf("expected cloned voice to be accelerated above baseline, got speed=%.2f atempo=%.2f", speedStd, atempoStd)
	}

	speedFast, atempoFast := tubie.ClonedVoiceTiming("cloned-fast")
	if speedFast*atempoFast <= speedStd*atempoStd {
		t.Fatalf("expected cloned-fast total pace (%.3f) > cloned pace (%.3f)", speedFast*atempoFast, speedStd*atempoStd)
	}

	filter := tubie.BuildClonedAudioFilter(atempoStd)
	if !strings.Contains(filter, "atempo=1.10") || !strings.Contains(filter, "loudnorm") {
		t.Fatalf("unexpected cloned audio filter chain: %q", filter)
	}

	// Verify FFmpeg accepts the filter chain offline using synthetic audio.
	ffmpeg := tubie.BinaryPath("ffmpeg")
	tmpDir := t.TempDir()
	rawWav := filepath.Join(tmpDir, "raw.wav")
	outWav := filepath.Join(tmpDir, "out.wav")

	genCmd := exec.Command(ffmpeg, "-v", "error", "-y", "-f", "lavfi", "-i", "sine=frequency=220:duration=2", "-ar", "24000", "-ac", "1", rawWav)
	if out, err := genCmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg synthetic wave generation failed: %v: %s", err, string(out))
	}

	procCmd := exec.Command(ffmpeg, "-v", "error", "-y", "-i", rawWav, "-af", filter, "-ar", "24000", "-ac", "1", outWav)
	if out, err := procCmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg cloned filter execution failed: %v: %s", err, string(out))
	}

	dur, err := tubie.GetAudioDuration(outWav)
	if err != nil {
		t.Fatalf("GetAudioDuration failed: %v", err)
	}
	if dur >= 1.95 || dur <= 1.0 {
		t.Fatalf("expected accelerated audio duration between 1.0s and 1.95s, got %.2fs", dur)
	}
}

func TestGenerateScriptStylesAndTopics(t *testing.T) {
	styles := []string{tubie.StyleHybrid, tubie.StyleCinematic, tubie.StyleInformative}
	seenNarrations := make(map[string]string)

	for _, style := range styles {
		res, err := tubie.GenerateScriptWithStyle("Sahte Linkleri 5 Saniyede Nasıl Anlarsın?", style)
		if err != nil {
			t.Fatalf("GenerateScriptWithStyle(%s) failed: %v", style, err)
		}
		if res.Style != style {
			t.Fatalf("expected style %q, got %q", style, res.Style)
		}
		if strings.Contains(strings.ToLower(res.Narration), "sandığından çok daha kritik") {
			t.Fatalf("cliche filler detected in %s script: %s", style, res.Narration)
		}
		if len(res.Scenes) != 6 {
			t.Fatalf("expected 6 scenes for %s, got %d", style, len(res.Scenes))
		}
		if prevStyle, dup := seenNarrations[res.Narration]; dup {
			t.Fatalf("styles %q and %q produced identical phishing narration", prevStyle, style)
		}
		seenNarrations[res.Narration] = style
	}

	// Dynamic unknown topic should also generate coherent, distinct non-boilerplate scripts.
	for _, topic := range []string{"Akıllı Ev Kameraları", "LLM Prompt Injection"} {
		hybrid, err := tubie.GenerateScriptWithStyle(topic, tubie.StyleHybrid)
		if err != nil {
			t.Fatalf("dynamic topic %q failed: %v", topic, err)
		}
		cinematic, err := tubie.GenerateScriptWithStyle(topic, tubie.StyleCinematic)
		if err != nil {
			t.Fatalf("dynamic cinematic %q failed: %v", topic, err)
		}
		if hybrid.Narration == cinematic.Narration {
			t.Fatalf("expected distinct narration across styles for topic %q", topic)
		}
		if hybrid.Slug == "" || strings.Contains(hybrid.Slug, " ") {
			t.Fatalf("invalid slug generated for %q: %q", topic, hybrid.Slug)
		}
		if len(hybrid.Scenes) != 6 {
			t.Fatalf("expected 6 scenes for dynamic topic %q, got %d", topic, len(hybrid.Scenes))
		}
	}
}

func TestGenerateScriptViaOllamaMock(t *testing.T) {
	mockPayload := `{
  "title": "BU LİNKE TIKLARSAN NE OLUR?",
  "narration": "Bir mesaj geliyor… Gönderen tanıdık gibi. Logo gerçek. Mesaj oldukça inandırıcı. Oysa tek tıkla oturumun çalınabilir. Önce alan adını kontrol et, acele ettiren baskıya kanma ve kilit simgesine güvenme. Tıklamadan önce dur, düşün, doğrula!",
  "scenes": [
    {"title": "TANIDIK MESAJ", "pose": "panic", "prop": "@b1/cyber-laptop.png", "sfx": "alert", "mood": "alert"},
    {"title": "SAHTE ALAN ADI", "pose": "present", "prop": "@b1/cyber-terminal.png", "sfx": "keyboard", "mood": "danger"},
    {"title": "ACİLİYET BASKISI", "pose": "panic", "prop": "@b1/cyber-skull.png", "sfx": "glitch", "mood": "danger"},
    {"title": "HTTPS TUZAĞI", "pose": "think", "prop": "@b1/cyber-padlock-unlocked.png", "sfx": "paper", "mood": "focus"},
    {"title": "BEŞ SANİYE TESTİ", "pose": "inspect", "prop": "@b1/idea-note-crayon.png", "sfx": "tap", "mood": "focus"},
    {"title": "DOĞRULA GÜVENDE KAL", "pose": "thumbsup", "prop": "@b1/cyber-shield.png", "sfx": "chime", "mood": "safe"}
  ]
}`

	var gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models":[{"name":"llama3.1:8b"},{"name":"qwen2.5:7b"}]}`))
		case "/api/generate":
			var req map[string]any
			_ = json.NewDecoder(r.Body).Decode(&req)
			gotModel, _ = req["model"].(string)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"response": "```json\n" + mockPayload + "\n```",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	res, err := tubie.GenerateScriptViaOllamaAt(srv.URL, "Sahte link tuzağı", tubie.StyleHybrid)
	if err != nil {
		t.Fatalf("GenerateScriptViaOllamaAt failed: %v", err)
	}
	if !strings.HasPrefix(gotModel, "qwen2.5:7b") {
		t.Fatalf("expected model selection to prefer qwen2.5:7b, got %q", gotModel)
	}
	if !strings.HasPrefix(res.Engine, "ollama:") {
		t.Fatalf("expected ollama engine tag, got %q", res.Engine)
	}
	if len(res.Scenes) != 6 {
		t.Fatalf("expected 6 normalized scenes, got %d", len(res.Scenes))
	}
}

func TestStudioWebAssetsAndAIGenerateHandler(t *testing.T) {
	root, err := tubie.FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}

	indexBytes, err := os.ReadFile(filepath.Join(root, "internal", "tubie", "web", "index.html"))
	if err != nil {
		t.Fatalf("failed reading index.html: %v", err)
	}
	indexHTML := string(indexBytes)
	if !strings.Contains(indexHTML, `id="ai-style"`) || !strings.Contains(indexHTML, `value="cloned-fast"`) {
		t.Fatalf("index.html missing #ai-style or cloned-fast option")
	}

	appJSBytes, err := os.ReadFile(filepath.Join(root, "internal", "tubie", "web", "app.js"))
	if err != nil {
		t.Fatalf("failed reading app.js: %v", err)
	}
	appJS := string(appJSBytes)
	if strings.Contains(appJS, `$('#mode')`) || !strings.Contains(appJS, `$('#voice-style')`) || !strings.Contains(appJS, `$('#ai-style')`) {
		t.Fatalf("app.js still references #mode or is missing #voice-style/#ai-style bindings")
	}

	studio, err := tubie.NewStudioServer(root, 0)
	if err != nil {
		t.Fatalf("NewStudioServer failed: %v", err)
	}
	defer studio.Close()
	go func() {
		_ = studio.Serve()
	}()

	baseURL := strings.Split(studio.URL(""), "#")[0]
	reqBody := bytes.NewBufferString(`{"prompt":"Sahte linkleri 5 saniyede nasıl anlarsın?","style":"cinematic"}`)
	resp, err := http.Post(baseURL+"api/ai-generate", "application/json", reqBody)
	if err != nil {
		t.Fatalf("POST /api/ai-generate failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/ai-generate, got %d", resp.StatusCode)
	}

	var decoded struct {
		OK     bool                `json:"ok"`
		Result *tubie.ScriptResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if !decoded.OK || decoded.Result == nil || decoded.Result.Style != tubie.StyleCinematic {
		t.Fatalf("unexpected api/ai-generate response: %+v", decoded)
	}
}

func TestAlignScriptWordsHandlesSplitsAndInsertions(t *testing.T) {
	canonical := tubie.ExtractScriptWords("Büyük tehlikenin habercisi olabilir. HTTPS tek başına garanti değildir. 5 saniyede anla, güvende kal!")
	// Recognized stream with a split word, a misheard acronym, an inserted filler, and a digit spoken as a word.
	recognized := []tubie.TimedWord{
		{Word: "büyük", Start: 0.10, End: 0.40},
		{Word: "tehlike", Start: 0.45, End: 0.80},
		{Word: "nin", Start: 0.80, End: 0.95},
		{Word: "habercisi", Start: 1.00, End: 1.50},
		{Word: "olabilir.", Start: 1.55, End: 2.00},
		{Word: "Hat", Start: 2.60, End: 2.80},
		{Word: "DPS", Start: 2.80, End: 3.10},
		{Word: "ıı", Start: 3.10, End: 3.20},
		{Word: "tek", Start: 3.25, End: 3.45},
		{Word: "başına", Start: 3.50, End: 3.90},
		{Word: "garanti", Start: 3.95, End: 4.40},
		{Word: "değildir.", Start: 4.45, End: 5.00},
		{Word: "beş", Start: 5.60, End: 5.80},
		{Word: "saniyede", Start: 5.85, End: 6.30},
		{Word: "anla,", Start: 6.35, End: 6.70},
		{Word: "güvende", Start: 6.80, End: 7.20},
		{Word: "kal.", Start: 7.25, End: 7.50},
	}

	timed := tubie.AlignScriptWords(canonical, recognized, nil, 8.0)
	if len(timed) != len(canonical) {
		t.Fatalf("expected %d timed words, got %d", len(canonical), len(timed))
	}
	expectStart := map[string]float64{"HABERCİSİ": 1.00, "TEK": 3.25, "GARANTİ": 3.95, "5": 5.60, "SANİYEDE": 5.85, "KAL": 7.25}
	for _, w := range timed {
		if want, ok := expectStart[w.Word]; ok && math.Abs(w.Start-want) > 0.01 {
			t.Errorf("word %s: expected start %.2f, got %.3f (drift)", w.Word, want, w.Start)
		}
	}
	for i := 1; i < len(timed); i++ {
		if timed[i].Start < timed[i-1].Start-1e-9 {
			t.Fatalf("non-monotonic timing at %d: %+v after %+v", i, timed[i], timed[i-1])
		}
	}
	if last := timed[len(timed)-1]; math.Abs(last.End-7.50) > 0.01 {
		t.Fatalf("expected final word to end with the speech (7.50s), got %.3f", last.End)
	}

	cues := tubie.GroupCaptionCues(timed, 8.0)
	prevEnd := 0.0
	for i, c := range cues {
		if c.Start < prevEnd || c.End <= c.Start {
			t.Fatalf("cue %d invalid/overlapping: %+v (prev end %.3f)", i, c, prevEnd)
		}
		prevEnd = c.End
	}
}

func TestAlignScriptWordsFallsBackToSpeechRegions(t *testing.T) {
	canonical := []string{"BİR", "MESAJ", "GELİYOR", "PEKİ", "YA", "TUZAKSA"}
	speech := []tubie.SpeechSegment{{Start: 0.5, End: 2.0}, {Start: 4.0, End: 5.5}}

	timed := tubie.AlignScriptWords(canonical, nil, speech, 6.0)
	if timed[0].Start < 0.5-1e-9 {
		t.Fatalf("first word starts before speech begins: %.3f", timed[0].Start)
	}
	if last := timed[len(timed)-1]; last.End > 5.5+1e-9 {
		t.Fatalf("last word ends after speech ends: %.3f", last.End)
	}
	for _, w := range timed {
		if w.Start > 2.0+1e-9 && w.Start < 4.0-1e-9 {
			t.Fatalf("word %s placed inside silence gap at %.3f", w.Word, w.Start)
		}
	}
}

func TestDetectSpeechSegmentsOffline(t *testing.T) {
	ffmpeg := tubie.BinaryPath("ffmpeg")
	wav := filepath.Join(t.TempDir(), "speech.wav")
	// 1s tone, 1s silence, 1s tone
	cmd := exec.Command(ffmpeg, "-v", "error", "-y",
		"-f", "lavfi", "-i", "sine=frequency=300:duration=1",
		"-f", "lavfi", "-i", "anullsrc=r=24000:cl=mono:d=1",
		"-f", "lavfi", "-i", "sine=frequency=300:duration=1",
		"-filter_complex", "[0:a]aresample=24000[a];[2:a]aresample=24000[c];[a][1:a][c]concat=n=3:v=0:a=1",
		"-ac", "1", wav)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("synthetic audio failed: %v: %s", err, string(out))
	}
	segs := tubie.DetectSpeechSegments(wav, 3.0)
	if len(segs) != 2 {
		t.Fatalf("expected 2 speech segments, got %d: %+v", len(segs), segs)
	}
	if math.Abs(segs[0].End-1.0) > 0.1 || math.Abs(segs[1].Start-2.0) > 0.1 {
		t.Fatalf("unexpected speech boundaries: %+v", segs)
	}
}

func TestAIPrivacyTopicDoesNotHijackToDeepfakeOrInstagram(t *testing.T) {
	// Exact user scenario: "Yapay zeka sizi internette ne kadar tanıyor?"
	res, err := tubie.GenerateScriptWithStyle("Yapay zeka sizi internette ne kadar tanıyor?", tubie.StyleHybrid)
	if err != nil {
		t.Fatalf("GenerateScriptWithStyle failed: %v", err)
	}
	lowNarr := strings.ToLower(res.Narration)
	if strings.Contains(lowNarr, "kaza yaptım") || strings.Contains(lowNarr, "sesini birebir klonlayabiliyor") {
		t.Fatalf("AI privacy prompt was hijacked to deepfake voice cloning: %s", res.Narration)
	}
	if !strings.Contains(lowNarr, "yapay zeka") {
		t.Fatalf("expected narration to stay on AI privacy topic, got: %s", res.Narration)
	}

	// Regression check for the exact output/yapay_zeka_sizi_internette_ne_kadar_tani/aciklama.txt bug:
	// Narration mentions "sosyal medyada paylaştıkların" and "sağlık sorunlarını" ("ağ" substring).
	userTitle := "Yapay Zeka Sizi İnternette Ne Kadar Tanıyor?"
	userNarration := "Yapay zeka seni internette ne kadar tanıyor? Belki de tahmin ettiğinden çok daha fazla. Çünkü artık sadece sosyal medyada paylaştıkların değil, yapay zekayla yaptığın sohbetler de senin hakkında bilgi veriyor. Ona hangi şehirde yaşadığını, nerede çalıştığını, ilişkini, korkularını, sağlık sorunlarını anlatıyor olabilirsin. Peki kendini nasıl koruyabilirsin? Bir: Yapay zekayla konuşurken şifre, kimlik numarası, adres ve banka bilgilerini paylaşma. İki: Özel fotoğraflarını, belgelerini ve başkalarına ait kişisel bilgileri yüklemeden önce iki kez düşün. Üç: Kullandığın yapay zeka araçlarının veri saklama, sohbet geçmişi ve model eğitimi ayarlarını kontrol et."

	vm := tubie.GenerateViralMetadata(userTitle, userNarration)
	combined := strings.ToLower(vm.TitleA + " " + vm.TitleB + " " + vm.TitleC + " " + vm.Description + " " + strings.Join(vm.Hashtags, " "))
	if strings.Contains(combined, "telif hakkı") || strings.Contains(combined, "instagram hesabınız") || strings.Contains(combined, "evil twin") {
		t.Fatalf("GenerateViralMetadata hijacked AI privacy video into unrelated template:\nTitleA: %s\nDesc: %s", vm.TitleA, vm.Description)
	}
	if !strings.Contains(combined, "yapay zeka") {
		t.Fatalf("expected metadata to reference yapay zeka, got: %s", combined)
	}
	if !strings.Contains(strings.ToLower(vm.Description), "veri saklama") && !strings.Contains(strings.ToLower(vm.Description), "kimlik numarası") {
		t.Fatalf("expected description to reflect actual narration rules, got: %s", vm.Description)
	}
	postCaption := vm.Description + "\n\n" + strings.Join(vm.Hashtags, " ")
	if len([]rune(postCaption)) > 2200 {
		t.Fatalf("caption exceeds 2200 runes: %d", len([]rune(postCaption)))
	}
}

func TestNormalizeViralMetadataSanitizesAndBounds(t *testing.T) {
	dirty := tubie.ViralMetadata{
		TitleA:      "Kuantum Şifreleme Nedir?",
		TitleB:      "Şifreleme Riski",
		TitleC:      "Korunma Rehberi",
		Description: strings.Repeat("Kuantum bilgisayarlar modern şifrelemeyi tehdit ediyor. ", 60),
		Hashtags:    []string{"#siberg%C3%BCvenlik", "kuantum", "#kuantum", "#bad tag!", "#yapayzeka"},
	}
	norm := tubie.NormalizeViralMetadata(dirty, "Kuantum Bilgisayarlar", "Kuantum bilgisayarlar ve şifreleme güvenliği.")
	if !strings.HasPrefix(norm.TitleA, "🚨") || !strings.HasPrefix(norm.TitleB, "⚠️") || !strings.HasPrefix(norm.TitleC, "🛡️") {
		t.Fatalf("missing required title emojis: %+v", norm)
	}
	joinedTags := strings.Join(norm.Hashtags, " ")
	if strings.Contains(joinedTags, "%") || strings.Contains(joinedTags, "bad tag") {
		t.Fatalf("invalid hashtag survived sanitization: %s", joinedTags)
	}
	if !strings.Contains(joinedTags, "#sibergüvenlik") {
		t.Fatalf("expected URL-encoded Turkish hashtag to decode cleanly, got: %s", joinedTags)
	}
	full := norm.Description + "\n\n" + joinedTags
	if len([]rune(full)) > 2200 {
		t.Fatalf("normalized metadata exceeds 2200 runes: %d", len([]rune(full)))
	}
}

func TestEnsureScenesForScriptAndStorytellingBounds(t *testing.T) {
	title := "Yapay Zeka Sizi İnternette Ne Kadar Tanıyor?"
	narration := "Yapay zeka seni internette ne kadar tanıyor? Belki de tahmin ettiğinden çok daha fazla. Sosyal medya dışında yapay zeka sohbetlerin de bilgi veriyor. Bir: Şifre ve kimlik bilgilerini paylaşma. İki: Özel belgeleri yüklemeden önce iki kez düşün. Üç: Sohbet geçmişi ve model eğitimi ayarlarını kontrol et. Özel hayatını koru!"
	scenes := tubie.EnsureScenesForScript(title, narration)
	if len(scenes) != 6 {
		t.Fatalf("expected 6 derived scenes, got %d", len(scenes))
	}
	for i, sc := range scenes {
		if sc.Title == "" || strings.HasPrefix(sc.Title, "SAHNE ") {
			t.Fatalf("unexpected generic/empty scene title at %d: %q", i, sc.Title)
		}
		if sc.Prop == nil || !strings.HasPrefix(sc.Prop.Path, "@b1/") {
			t.Fatalf("invalid scene prop at %d: %+v", i, sc.Prop)
		}
	}
	shots := tubie.BuildStoryboard(scenes, 45.0)
	if len(shots) != 6 {
		t.Fatalf("expected 6 storyboard shots, got %d", len(shots))
	}
}

func TestDescribeSubcommand(t *testing.T) {
	if code := tubie.Main([]string{"describe"}); code == 0 {
		t.Fatalf("expected non-zero exit when --text and --title are missing")
	}

	outPath := filepath.Join(t.TempDir(), "nested", "aciklama.txt")
	code := tubie.Main([]string{
		"describe",
		"--title", "Yapay Zeka Sizi İnternette Ne Kadar Tanıyor?",
		"--text", "Yapay zeka sohbetlerin sağlık sorunlarını ve kişisel bilgilerini ortaya çıkarabilir. Veri saklama ve model eğitimi ayarlarını kontrol et.",
		"--out", outPath,
	})
	if code != 0 {
		t.Fatalf("describe subcommand failed with exit code %d", code)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed reading written aciklama.txt: %v", err)
	}
	content := string(data)
	for _, want := range []string{"[BAŞLIK A", "[BAŞLIK B", "[BAŞLIK C", "YOUTUBE TAG"} {
		if !strings.Contains(content, want) {
			t.Fatalf("missing section %q in describe output:\n%s", want, content)
		}
	}
	if strings.Contains(strings.ToLower(content), "telif hakkı") {
		t.Fatalf("describe output hijacked into unrelated template:\n%s", content)
	}
}

func TestViralAnalyzerScoring(t *testing.T) {
	ideas := tubie.TrendingCyberIdeas()
	if len(ideas) != 18 {
		t.Fatalf("expected 18 curated cyber ideas, got %d", len(ideas))
	}
	for _, idea := range ideas {
		if idea.ID == "" || idea.Category == "" || idea.Title == "" || idea.Prompt == "" {
			t.Fatalf("incomplete curated cyber idea: %+v", idea)
		}
	}

	if cat := tubie.DetectThreatCategory("Kahvecide ücretsiz Wi-Fi Deauth ve Evil Twin saldırısı"); !strings.Contains(cat, "Wi-Fi") {
		t.Fatalf("unexpected category for Wi-Fi topic: %s", cat)
	}
	if cat := tubie.DetectThreatCategory("3 saniyelik videodan yapay zeka deepfake ses klonlama"); !strings.Contains(cat, "Yapay Zeka") {
		t.Fatalf("unexpected category for AI topic: %s", cat)
	}

	strongTitle := "3 Saniyede Wi-Fi Şifren Nasıl Çalınır?"
	strongText := "Şu an bağlandığın kafe Wi-Fi ağı senin tüm şifrelerini çalabilir! Hacker kaba kuvvet denemez, tek paketle seni ağdan düşürür. Oysa bağlandığın anda el sıkışma paketi havada yakalanır ve %95 risk doğar! Kendini 2 adımda koru: 1. Asla şifresiz ortak ağa güvenme. 2. Hemen VPN aç ve 2FA doğrulamayı etkinleştir. Bu videoyu kaydet ve paylaş; işte bu yüzden:"
	strongScenes := tubie.EnsureScenesForScript(strongTitle, strongText)
	strong := tubie.AnalyzeScriptVirality(strongTitle, strongText, strongScenes, true)

	weak := tubie.AnalyzeScriptVirality("Video", "Merhaba bugün biraz internetten konuşalım.", nil, false)
	if strong.TotalScore <= weak.TotalScore+25 {
		t.Fatalf("expected strong script (%d) to beat weak script (%d) by >25 points", strong.TotalScore, weak.TotalScore)
	}
	if len(strong.Dimensions) != 6 {
		t.Fatalf("expected 6 viral dimensions, got %d", len(strong.Dimensions))
	}
	dimSum := 0
	for _, d := range strong.Dimensions {
		dimSum += d.Score
	}
	if dimSum != strong.TotalScore {
		t.Fatalf("dimension sum %d != TotalScore %d", dimSum, strong.TotalScore)
	}
	if !strong.LoopReady || weak.LoopReady {
		t.Fatalf("expected strong.LoopReady=true and weak.LoopReady=false, got strong=%v weak=%v", strong.LoopReady, weak.LoopReady)
	}
}

func TestAllSixViralStylesAndLoopMode(t *testing.T) {
	t.Setenv("TUBIE_OFFLINE_AI", "1")
	styles := []string{
		tubie.StyleHybrid,
		tubie.StyleCinematic,
		tubie.StyleInformative,
		tubie.StyleMythBuster,
		tubie.StyleChallenge,
		tubie.StyleRedVsBlue,
	}
	seenNarr := make(map[string]string)
	for _, st := range styles {
		res, err := tubie.GenerateScriptWithOptions("Sahte banka SMS ve oltalama linki", tubie.ScriptOptions{
			Style:      st,
			LoopEnding: false,
		})
		if err != nil {
			t.Fatalf("style %s failed: %v", st, err)
		}
		if res.Style != st {
			t.Fatalf("expected style %s, got %s", st, res.Style)
		}
		if len(res.Scenes) != 6 {
			t.Fatalf("style %s: expected 6 scenes, got %d", st, len(res.Scenes))
		}
		if res.ViralScore == nil || res.ViralScore.TotalScore <= 0 {
			t.Fatalf("style %s: expected non-zero ViralScore, got %+v", st, res.ViralScore)
		}
		if prevStyle, dup := seenNarr[res.Narration]; dup {
			t.Fatalf("style %s produced identical narration to %s", st, prevStyle)
		}
		seenNarr[res.Narration] = st
	}

	loopRes, err := tubie.GenerateScriptWithOptions("Sahte banka SMS ve oltalama linki", tubie.ScriptOptions{
		Style:      tubie.StyleHybrid,
		LoopEnding: true,
	})
	if err != nil {
		t.Fatalf("loop generation failed: %v", err)
	}
	if !loopRes.LoopEnding || loopRes.ViralScore == nil || !loopRes.ViralScore.LoopReady {
		t.Fatalf("expected loop metadata on generated result, got loop=%v score=%+v", loopRes.LoopEnding, loopRes.ViralScore)
	}
}

func TestSyncedStoryboardPortraitAndLandscape(t *testing.T) {
	narration := "Bu linke sakın tıklama! Hacker kopya siteyle şifreni alır. Hemen göndericiyi kontrol et. Resmi uygulamadan giriş yap. Authenticator ile iki adımlı doğrulama aç. Tuzağa düşme ve güvende kal!"
	words := tubie.ExtractScriptWords(narration)
	timed := make([]tubie.TimedWord, len(words))
	step := 18.0 / float64(len(words))
	for i, w := range words {
		timed[i] = tubie.TimedWord{
			Word:  w,
			Start: math.Round(float64(i)*step*1000) / 1000,
			End:   math.Round((float64(i)*step+step*0.85)*1000) / 1000,
		}
	}
	scenes := tubie.EnsureScenesForScript("Sahte Link Tuzağı", narration)

	bounds := tubie.ComputeSceneBoundaries(scenes, timed, narration, 18.0)
	if len(bounds) != len(scenes)+1 || bounds[0] != 0 || bounds[len(bounds)-1] != 18.0 {
		t.Fatalf("unexpected scene boundaries: %v", bounds)
	}
	for i := 1; i < len(bounds); i++ {
		if bounds[i] <= bounds[i-1] {
			t.Fatalf("non-monotonic boundaries at %d: %v", i, bounds)
		}
	}

	for _, tc := range []struct {
		name string
		w, h int
	}{
		{"portrait", 720, 1280},
		{"landscape", 1280, 720},
	} {
		shots := tubie.BuildStoryboardWithOptions(scenes, 18.0, tubie.StoryboardOptions{
			Width:      tc.w,
			Height:     tc.h,
			TimedWords: timed,
			Narration:  narration,
		})
		if len(shots) != 6 {
			t.Fatalf("%s: expected 6 shots, got %d", tc.name, len(shots))
		}
		var sawActor, sawWidget, sawCallout bool
		for _, sh := range shots {
			for _, l := range sh["layers"].([]map[string]any) {
				switch l["type"] {
				case "cyber_actor":
					sawActor = true
				case "cyber_widget":
					sawWidget = true
				case "cyber_callout":
					sawCallout = true
				}
			}
		}
		if !sawActor || !sawWidget || !sawCallout {
			t.Fatalf("%s: missing cyber layers actor=%v widget=%v callout=%v", tc.name, sawActor, sawWidget, sawCallout)
		}
	}
}

func TestVoiceViralFilterOffline(t *testing.T) {
	dir := t.TempDir()
	rawWav := filepath.Join(dir, "raw.wav")
	outWav := filepath.Join(dir, "filtered.wav")
	ffmpeg := tubie.BinaryPath("ffmpeg")

	makeCmd := exec.Command(ffmpeg, "-v", "error", "-y",
		"-f", "lavfi", "-i", "anullsrc=r=24000:cl=mono:d=0.35",
		"-f", "lavfi", "-i", "sine=f=320:r=24000:d=1.0",
		"-filter_complex", "[0:a][1:a]concat=n=2:v=0:a=1",
		rawWav)
	if out, err := makeCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed creating synthetic wav: %v (%s)", err, string(out))
	}

	filter := tubie.BuildViralVocalFilter(1.06)
	if !strings.Contains(filter, "silenceremove=") || !strings.Contains(filter, "loudnorm=") {
		t.Fatalf("unexpected viral vocal filter chain: %s", filter)
	}
	filtCmd := exec.Command(ffmpeg, "-v", "error", "-y", "-i", rawWav, "-af", filter, "-ar", "24000", "-ac", "1", outWav)
	if out, err := filtCmd.CombinedOutput(); err != nil {
		t.Fatalf("viral vocal filter execution failed: %v (%s)", err, string(out))
	}

	dur, err := tubie.GetAudioDuration(outWav)
	if err != nil {
		t.Fatalf("GetAudioDuration failed: %v", err)
	}
	if dur <= 0.4 || dur >= 1.25 {
		t.Fatalf("expected tightened & tempo-adjusted duration in (0.4, 1.25), got %.3f", dur)
	}
}

func TestDualFormatAndABThumbnails(t *testing.T) {
	root, err := tubie.FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}
	dir := t.TempDir()
	thumb9A := filepath.Join(dir, "thumb_9x16_a.jpg")
	thumb9B := filepath.Join(dir, "thumb_9x16_b.jpg")
	thumb16A := filepath.Join(dir, "thumb_16x9_a.jpg")

	for _, job := range []struct {
		out  string
		opts tubie.ThumbnailOptions
	}{
		{thumb9A, tubie.ThumbnailOptions{Aspect: "9:16", Variant: "a", Mood: "danger"}},
		{thumb9B, tubie.ThumbnailOptions{Aspect: "9:16", Variant: "b", Mood: "defense"}},
		{thumb16A, tubie.ThumbnailOptions{Aspect: "16:9", Variant: "a", Mood: "breach"}},
	} {
		if err := tubie.GenerateDedicatedThumbnailWithOptions(root, "Bu Linke Tıklarsan Ne Olur?", "Sahte URL oltalama", "@b1/cyber-skull.png", 2, job.opts, job.out, ""); err != nil {
			t.Fatalf("thumbnail %+v failed: %v", job.opts, err)
		}
	}

	ffprobe := tubie.BinaryPath("ffprobe")
	readDims := func(path string) (int, int) {
		out, err := exec.Command(ffprobe, "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "json", path).Output()
		if err != nil {
			t.Fatalf("ffprobe failed on %s: %v", path, err)
		}
		var parsed struct {
			Streams []struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			} `json:"streams"`
		}
		if err := json.Unmarshal(out, &parsed); err != nil || len(parsed.Streams) == 0 {
			t.Fatalf("failed decoding probe on %s: %v (%s)", path, err, string(out))
		}
		return parsed.Streams[0].Width, parsed.Streams[0].Height
	}

	w9, h9 := readDims(thumb9A)
	if w9 >= h9 || math.Abs(float64(w9)/float64(h9)-9.0/16.0) > 0.02 {
		t.Fatalf("expected 9:16 portrait dimensions, got %dx%d", w9, h9)
	}
	w16, h16 := readDims(thumb16A)
	if w16 <= h16 || math.Abs(float64(w16)/float64(h16)-16.0/9.0) > 0.02 {
		t.Fatalf("expected 16:9 landscape dimensions, got %dx%d", w16, h16)
	}

	bytesA, _ := os.ReadFile(thumb9A)
	bytesB, _ := os.ReadFile(thumb9B)
	if bytes.Equal(bytesA, bytesB) {
		t.Fatalf("expected A and B thumbnail variants to differ visually")
	}
}

func TestTurkishCaptionSingleLineSafety(t *testing.T) {
	words := tubie.ExtractScriptWords("İSTANBUL içindeki kullanıcıların ŞİFRELERİNİZİ ve DOĞRULAMA kodlarını siber güvenlik uzmanları korur!")
	timed := make([]tubie.TimedWord, len(words))
	for i, w := range words {
		timed[i] = tubie.TimedWord{
			Word:  w,
			Start: float64(i) * 0.45,
			End:   float64(i)*0.45 + 0.38,
		}
	}
	cues := tubie.GroupCaptionCues(timed, 10.0)
	if len(cues) == 0 {
		t.Fatal("expected non-empty Turkish cues")
	}
	for i, c := range cues {
		if strings.ContainsAny(c.Text, "\r\n") {
			t.Fatalf("cue %d contains newline: %q", i, c.Text)
		}
		if len([]rune(c.Text)) > 26 {
			t.Fatalf("cue %d unexpectedly wide (%d runes): %q", i, len([]rune(c.Text)), c.Text)
		}
	}
	vtt := tubie.FormatVTT(cues)
	if !strings.HasPrefix(vtt, "WEBVTT\n\n") || !strings.Contains(vtt, ".") {
		t.Fatalf("invalid VTT output:\n%s", vtt)
	}
}

func TestStudioServerSecurityAndViralEndpoints(t *testing.T) {
	root, err := tubie.FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(root, "projects", "_thumb_preview"))
	})

	studio, err := tubie.NewStudioServer(root, 0)
	if err != nil {
		t.Fatalf("NewStudioServer failed: %v", err)
	}
	defer studio.Close()
	go func() {
		_ = studio.Serve()
	}()

	baseURL := strings.Split(studio.URL(""), "#")[0]

	// Mutating / POST-only endpoints must reject GET with 405 Method Not Allowed.
	for _, ep := range []string{"api/approve", "api/cancel", "api/shutdown", "api/viral-analyze", "api/thumbnail-preview"} {
		resp, err := http.Get(baseURL + ep)
		if err != nil {
			t.Fatalf("GET %s failed: %v", ep, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("GET %s: expected 405, got %d", ep, resp.StatusCode)
		}
	}

	// Path traversal / restricted file check on /media/.
	travResp, err := http.Get(baseURL + "media/output/../../go.mod")
	if err != nil {
		t.Fatalf("GET traversal check failed: %v", err)
	}
	travResp.Body.Close()
	if travResp.StatusCode != http.StatusBadRequest && travResp.StatusCode != http.StatusForbidden && travResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 400/403/404 on path traversal, got %d", travResp.StatusCode)
	}
	secretResp, err := http.Get(baseURL + "media/projects/foo/output/project.json")
	if err != nil {
		t.Fatalf("GET restricted project file failed: %v", err)
	}
	secretResp.Body.Close()
	if secretResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden on projects/.../project.json, got %d", secretResp.StatusCode)
	}

	// GET /api/ideas
	ideasResp, err := http.Get(baseURL + "api/ideas")
	if err != nil {
		t.Fatalf("GET api/ideas failed: %v", err)
	}
	defer ideasResp.Body.Close()
	var ideasPayload struct {
		OK    bool              `json:"ok"`
		Ideas []tubie.CyberIdea `json:"ideas"`
	}
	if err := json.NewDecoder(ideasResp.Body).Decode(&ideasPayload); err != nil || !ideasPayload.OK || len(ideasPayload.Ideas) != 18 {
		t.Fatalf("unexpected api/ideas response: %+v err=%v", ideasPayload, err)
	}

	// POST /api/viral-analyze
	analyzeResp, err := http.Post(baseURL+"api/viral-analyze", "application/json",
		bytes.NewBufferString(`{"title":"Bu Linke Tıklama!","narration":"3 saniyede şifren çalınır! Hemen 2FA aç ve bu videoyu kaydet.","loop_ending":true}`))
	if err != nil {
		t.Fatalf("POST api/viral-analyze failed: %v", err)
	}
	defer analyzeResp.Body.Close()
	var analyzePayload struct {
		OK     bool                    `json:"ok"`
		Report *tubie.ViralScoreReport `json:"report"`
	}
	if err := json.NewDecoder(analyzeResp.Body).Decode(&analyzePayload); err != nil || !analyzePayload.OK || analyzePayload.Report == nil || len(analyzePayload.Report.Dimensions) != 6 {
		t.Fatalf("unexpected api/viral-analyze response: %+v err=%v", analyzePayload, err)
	}

	// POST /api/thumbnail-preview and fetch generated JPEG via /media/.
	thumbResp, err := http.Post(baseURL+"api/thumbnail-preview", "application/json",
		bytes.NewBufferString(`{"title":"Sahte Link Tuzağı","topic":"Phishing","mood":"danger"}`))
	if err != nil {
		t.Fatalf("POST api/thumbnail-preview failed: %v", err)
	}
	defer thumbResp.Body.Close()
	var thumbPayload struct {
		OK          bool   `json:"ok"`
		Preview9x16 string `json:"preview_9x16"`
		Preview16x9 string `json:"preview_16x9"`
	}
	if err := json.NewDecoder(thumbResp.Body).Decode(&thumbPayload); err != nil || !thumbPayload.OK || thumbPayload.Preview9x16 == "" || thumbPayload.Preview16x9 == "" {
		t.Fatalf("unexpected api/thumbnail-preview response: %+v err=%v", thumbPayload, err)
	}
	imgResp, err := http.Get(baseURL + "media/" + thumbPayload.Preview9x16)
	if err != nil {
		t.Fatalf("GET preview thumbnail failed: %v", err)
	}
	imgResp.Body.Close()
	if imgResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for generated preview thumbnail, got %d", imgResp.StatusCode)
	}
}

func TestAnalyzeSubcommand(t *testing.T) {
	if code := tubie.Main([]string{"analyze"}); code == 0 {
		t.Fatalf("expected non-zero exit when --text and --title are missing")
	}
	code := tubie.Main([]string{
		"analyze",
		"--title", "Sahte Link Tuzağı",
		"--text", "Tıklama! 3 saniyede şifren çalınır. Hemen 2FA aç ve kaydet!",
		"--loop",
		"--json",
	})
	if code != 0 {
		t.Fatalf("analyze subcommand failed with exit code %d", code)
	}
}

func TestTurkishScriptQualityGateRejectsBrokenLLMOutput(t *testing.T) {
	brokenPollinations := `Şu an akıllı frizlatma cihazı çekiyor, ama gizli bir siber saldırı nereye gider uzak tutuşsa? O, tek bir uzaktan yönlendirilebilmiş kablosuz anahtarla başlar. Bir: Wi-Fi şifrenizi her 90 günde bir değiştirip güçlü karmaşık hal getirin. Efsane: Bir akıllı termostat, zaten 2019'da mikrovirus cihazını çaldığında 30 milyon akıllı sigortacı ders aldı. 70 % saldırıyı engeller. Şu an akıllı frizlatma cihazı çekiyor, ama gizli bir siber saldırı nereye gider uzak tutuşsa?`
	if err := tubie.ValidateTurkishScriptQuality("Akıllı Cihazlar", brokenPollinations, nil); err == nil {
		t.Fatalf("expected ValidateTurkishScriptQuality to reject broken Pollinations narration")
	}

	samples := []struct {
		name      string
		narration string
	}{
		{"meta-label", "Kanca: Telefonuna bildirim düştü. Bir: Şifreni hemen değiştir. İki: 2FA aç."},
		{"cyrillic", "Telefonunda gizli bir güvenlik открыта açığı tespit edildi."},
		{"spaced-pct", "Bu basit kural saldırıların 95 % oranında önüne geçer."},
		{"english-flood", "Your smart camera is exposed with weak password and this attack will steal your data."},
		{"duplicate", "Evindeki akıllı kamera şu an seni gizlice izliyor olabilir. Hemen admin şifresini değiştir. Misafir ağına geçir. Evindeki akıllı kamera şu an seni gizlice izliyor olabilir."},
	}
	for _, s := range samples {
		if err := tubie.ValidateTurkishScriptQuality("IoT Güvenliği", s.narration, nil); err == nil {
			t.Errorf("sample %q should have failed ValidateTurkishScriptQuality", s.name)
		}
	}

	clean := "Evindeki akıllı kamera şu an gizlice seni izliyor olabilir! Bir: Varsayılan admin şifresini kurulumda değiştir. İki: Akıllı cihazları misafir Wi-Fi ağına izole et. Üç: Otomatik firmware güncellemelerini aç ve kaydet!"
	if err := tubie.ValidateTurkishScriptQuality("IoT Cihaz Riskleri", clean, nil); err != nil {
		t.Fatalf("expected clean Turkish script to pass, got error: %v", err)
	}
}

func TestIoTTopicGenerationAndLoopBridge(t *testing.T) {
	t.Setenv("TUBIE_OFFLINE_AI", "1")
	for _, style := range []string{tubie.StyleHybrid, tubie.StyleCinematic, tubie.StyleRedVsBlue} {
		res, err := tubie.GenerateScriptWithOptions("IoT Cihazlarının Siber Güvenlik Riskleri", tubie.ScriptOptions{
			Style:      style,
			LoopEnding: true,
		})
		if err != nil {
			t.Fatalf("style %s failed: %v", style, err)
		}
		if res.ViralScore == nil {
			t.Fatalf("style %s returned nil ViralScore", style)
		}
		if res.ViralScore.ThreatCategory != "Ağ, Wi-Fi & Kablosuz" {
			t.Fatalf("style %s: expected category Ağ, Wi-Fi & Kablosuz, got %q", style, res.ViralScore.ThreatCategory)
		}
		if res.ViralScore.TotalScore < 82 {
			t.Fatalf("style %s: expected ViralScore >= 82, got %d (%+v)", style, res.ViralScore.TotalScore, res.ViralScore)
		}
		if !res.ViralScore.LoopReady {
			t.Fatalf("style %s: expected LoopReady=true, got narration=%q", style, res.Narration)
		}
		if strings.Contains(res.Narration, "tıklamadan önce") {
			t.Fatalf("style %s: IoT loop ending should not say 'tıklamadan önce': %q", style, res.Narration)
		}
		if strings.Contains(strings.ToLower(res.Narration), "efsane:") {
			t.Fatalf("style %s: unexpected meta label in narration: %q", style, res.Narration)
		}
		if len(res.ViralScore.Dimensions) >= 5 && res.ViralScore.Dimensions[4].Score < 10 {
			t.Fatalf("style %s: expected Saveability score >= 10 with Unicode Bir/İki/Üç parser, got %+v", style, res.ViralScore.Dimensions[4])
		}
	}
}

func TestHybridAITopicDoesNotHijackToStaticAIPrivacy(t *testing.T) {
	t.Setenv("TUBIE_OFFLINE_AI", "1")
	prompt := "Yapay Zekâ Destekli IoT Cihazlarında Siber Tehdit"
	res, err := tubie.GenerateScriptWithOptions(prompt, tubie.ScriptOptions{
		Style:      tubie.StyleHybrid,
		LoopEnding: true,
	})
	if err != nil {
		t.Fatalf("GenerateScriptWithOptions failed: %v", err)
	}
	if strings.Contains(strings.ToLower(res.Title), "ne kadar tanıyor") {
		t.Fatalf("hybrid AI+IoT prompt was hijacked to static AI-privacy template: title=%q", res.Title)
	}
	lowNarr := strings.ToLower(res.Narration)
	if !strings.Contains(lowNarr, "iot") || !strings.Contains(lowNarr, "yapay zek") {
		t.Fatalf("expected hybrid AI+IoT narration to cover both AI and IoT, got: %s", res.Narration)
	}
	if err := tubie.ValidateTurkishScriptQuality(res.Title, res.Narration, res.Scenes); err != nil {
		t.Fatalf("generated script failed quality gate: %v", err)
	}
	if !strings.Contains(res.Title, "IoT") {
		t.Fatalf("expected title to preserve canonical acronym IoT, got %q", res.Title)
	}
	if !strings.Contains(res.Slug, "zeka") {
		t.Fatalf("expected slug to normalize circumflex â -> a (zeka), got %q", res.Slug)
	}
	if res.ViralScore == nil || !utf8.ValidString(res.ViralScore.HookSentence) || strings.ContainsRune(res.ViralScore.HookSentence, '\uFFFD') {
		t.Fatalf("corrupted UTF-8 in HookSentence: %+v", res.ViralScore)
	}
	for i, sc := range res.Scenes {
		if sc.Callout == "SAHTE BİLDİRİM!" || sc.Callout == "HARFE DİKKAT!" {
			t.Fatalf("scene %d used phishing-only callout %q on AI+IoT topic", i, sc.Callout)
		}
		raw, _ := json.Marshal(sc.VisualData)
		s := strings.ToLower(string(raw))
		if strings.Contains(s, "hesap-dogru1ama.com") || strings.Contains(s, "15 dk içinde doğrulayın") || strings.Contains(s, "panikle tıkla") {
			t.Fatalf("scene %d leaked hardcoded phishing visual_data onto AI+IoT topic: %s", i, string(raw))
		}
	}

	// Regression check: multi-byte ellipsis '…' in first sentence must never slice mid-rune.
	ellipsisReport := tubie.AnalyzeScriptVirality("Test", "Evindeki cihaz seni dinliyor… İkinci cümle burada.", nil, false)
	if !utf8.ValidString(ellipsisReport.HookSentence) || strings.ContainsRune(ellipsisReport.HookSentence, '\uFFFD') {
		t.Fatalf("firstSentenceOf corrupted UTF-8 ellipsis: %q", ellipsisReport.HookSentence)
	}
	if !strings.HasSuffix(ellipsisReport.HookSentence, "…") {
		t.Fatalf("expected HookSentence to preserve full ellipsis rune, got %q", ellipsisReport.HookSentence)
	}
}

func TestLoopEndingStripsTrailingConjunctionAndEcho(t *testing.T) {
	t.Setenv("TUBIE_OFFLINE_AI", "1")
	res, err := tubie.GenerateScriptWithOptions("IoT Cihazlarının Siber Güvenlik Riskleri", tubie.ScriptOptions{
		Style:      tubie.StyleHybrid,
		LoopEnding: true,
	})
	if err != nil {
		t.Fatalf("failed generating IoT loop script: %v", err)
	}
	if strings.Contains(res.Narration, "ve... İşte") || strings.Contains(res.Narration, "ve… İşte") {
		t.Fatalf("dangling conjunction before loop bridge: %q", res.Narration)
	}
	// Also verify viral metadata on custom AI+IoT video never hijacks to unrelated static description
	vm := tubie.GenerateViralMetadata(
		"Yapay Zekâ Destekli IoT Cihazlarında Siber Tehdit",
		"Evindeki akıllı kamera ve sesli asistan internete bağlı. Bir: Varsayılan admin şifresini değiştir. İki: Akıllı cihazları misafir ağına izole et. Üç: Firmware güncellemelerini aç ve UPnP kapat.",
	)
	if strings.Contains(strings.ToLower(vm.Description), "instagram") || strings.Contains(strings.ToLower(vm.Description), "telif hakkı") {
		t.Fatalf("viral metadata hijacked custom IoT video: %s", vm.Description)
	}
	if !strings.Contains(strings.ToLower(vm.Description), "misafir") {
		t.Fatalf("expected viral metadata description to reflect actual narration rules, got: %s", vm.Description)
	}
}

func TestExtractSpeechLevelsAndWAVSegmentsOffline(t *testing.T) {
	dir := t.TempDir()
	wav := filepath.Join(dir, "speech_pattern.wav")
	ffmpeg := tubie.BinaryPath("ffmpeg")

	cmd := exec.Command(ffmpeg, "-v", "error", "-y",
		"-f", "lavfi", "-i", "anullsrc=r=24000:cl=mono:d=0.6",
		"-f", "lavfi", "-i", "sine=f=320:r=24000:d=1.1",
		"-f", "lavfi", "-i", "anullsrc=r=24000:cl=mono:d=0.8",
		"-filter_complex", "[0:a][1:a][2:a]concat=n=3:v=0:a=1",
		"-ar", "24000", "-ac", "1", wav,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed generating synthetic WAV: %v (%s)", err, string(out))
	}

	levels := tubie.ExtractSpeechLevels(wav, 30, 2.5)
	if len(levels) < 70 || len(levels) > 80 {
		t.Fatalf("expected ~75 speech level frames for 2.5s@30fps, got %d", len(levels))
	}
	for i, lv := range levels {
		if lv < 0 || lv > 1 {
			t.Fatalf("frame %d level out of [0,1]: %f", i, lv)
		}
	}
	for i := 0; i <= 12; i++ {
		if levels[i] >= 0.1 {
			t.Fatalf("expected quiet opening frame %d < 0.1, got %f", i, levels[i])
		}
	}
	peakMid := 0.0
	for i := 22; i <= 48; i++ {
		if levels[i] > peakMid {
			peakMid = levels[i]
		}
	}
	if peakMid < 0.6 {
		t.Fatalf("expected strong voiced middle peak > 0.6, got %f", peakMid)
	}

	segs := tubie.DetectSpeechSegmentsFromWAV(wav, 2.5)
	if len(segs) != 1 {
		t.Fatalf("expected 1 voiced speech segment, got %d: %+v", len(segs), segs)
	}
	if segs[0].Start < 0.45 || segs[0].Start > 0.72 || segs[0].End < 1.55 || segs[0].End > 1.85 {
		t.Fatalf("unexpected speech segment bounds: %+v", segs[0])
	}
}

func TestTurkishAcronymCasingAndSemanticCaptionBreaks(t *testing.T) {
	words, strongBreaks, _ := tubie.ExtractScriptTokens("AI destekli Wi-Fi phishing saldırısı! İstanbul'da şifrenizi koruyun?")
	want := []string{"AI", "DESTEKLİ", "WI-FI", "PHISHING", "SALDIRISI", "İSTANBULDA", "ŞİFRENİZİ", "KORUYUN"}
	if len(words) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(words), words)
	}
	for i := range want {
		if words[i] != want[i] {
			t.Errorf("token %d: want %q, got %q", i, want[i], words[i])
		}
	}
	if !strongBreaks[4] || !strongBreaks[7] || strongBreaks[0] {
		t.Fatalf("unexpected strongBreaks flags: %v", strongBreaks)
	}

	timed := make([]tubie.TimedWord, len(words))
	for i, w := range words {
		timed[i] = tubie.TimedWord{
			Word:       w,
			Start:      float64(i) * 0.25,
			End:        float64(i)*0.25 + 0.24,
			BreakAfter: strongBreaks[i],
		}
	}
	cues := tubie.GroupCaptionCues(timed, 5.0)
	for _, c := range cues {
		if strings.Contains(c.Text, "SALDIRISI") && strings.Contains(c.Text, "İSTANBULDA") {
			t.Fatalf("caption cue illegally merged across sentence boundary: %q", c.Text)
		}
	}

	commaWords, commaStrong, commaSoft := tubie.ExtractScriptTokens("HAYIR DUR, BAK!")
	commaTimed := make([]tubie.TimedWord, len(commaWords))
	for i, w := range commaWords {
		commaTimed[i] = tubie.TimedWord{
			Word:       w,
			Start:      float64(i) * 0.22,
			End:        float64(i)*0.22 + 0.21,
			BreakAfter: commaStrong[i],
			SoftBreak:  commaSoft[i],
		}
	}
	commaCues := tubie.GroupCaptionCues(commaTimed, 2.0)
	if len(commaCues) < 2 || commaCues[0].Text != "HAYIR DUR" {
		t.Fatalf("expected comma SoftBreak on 2nd word to split 3-word chunk into 'HAYIR DUR', got: %+v", commaCues)
	}
}

func TestDoctorWithLoggerStreamsLines(t *testing.T) {
	root, err := tubie.FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}
	var lines []string
	code := tubie.DoctorWithLogger(root, func(line string) {
		lines = append(lines, line)
	})
	if code != 0 {
		t.Fatalf("DoctorWithLogger returned non-zero code %d:\n%s", code, strings.Join(lines, "\n"))
	}
	joined := strings.Join(lines, "\n")
	for _, needle := range []string{"OK node", "OK ffmpeg", "OK ffprobe", "OK node module @napi-rs/canvas", "OK node module sharp"} {
		if !strings.Contains(joined, needle) {
			t.Errorf("expected DoctorWithLogger output to contain %q, got:\n%s", needle, joined)
		}
	}
}

func TestStudioServerHardeningAndPinnedComment(t *testing.T) {
	dir := t.TempDir()
	studio, err := tubie.NewStudioServer(dir, 0)
	if err != nil {
		t.Fatalf("NewStudioServer failed: %v", err)
	}
	defer studio.Close()
	go func() {
		_ = studio.Serve()
	}()

	baseURL := strings.Split(studio.URL(""), "#")[0]

	rootResp, err := http.Get(baseURL)
	if err != nil {
		t.Fatalf("GET root failed: %v", err)
	}
	rootResp.Body.Close()
	csp := rootResp.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'self'") || strings.Contains(csp, "script-src 'self' 'unsafe-inline'") {
		t.Fatalf("unexpected script-src in CSP header: %q", csp)
	}

	const overLimit = 50*1024*1024 + 1024
	oversize := make([]byte, overLimit)
	uploadResp, err := http.Post(baseURL+"api/voice", "application/octet-stream", bytes.NewReader(oversize))
	if err != nil {
		t.Fatalf("POST oversize api/voice failed: %v", err)
	}
	_, _ = io.ReadAll(uploadResp.Body)
	uploadResp.Body.Close()
	if uploadResp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 Request Entity Too Large for >50MB upload, got %d", uploadResp.StatusCode)
	}

	vm := tubie.GenerateViralMetadata("Wi-Fi Tuzakları", "Ortak Wi-Fi ağına bağlanma. VPN aç ve şifreni koru.")
	formatted := tubie.FormatViralText("Wi-Fi Tuzakları", "Ortak Wi-Fi ağına bağlanma. VPN aç ve şifreni koru.", vm)
	if !strings.Contains(formatted, "SABİTLENECEK YORUM KANCASI") {
		t.Fatalf("expected FormatViralText to include 'SABİTLENECEK YORUM KANCASI', got:\n%s", formatted)
	}

	for _, p := range tubie.PresetList() {
		if p.Title == "" || p.Narration == "" || len(p.Scenes) != 6 {
			t.Errorf("preset %s invalid (title=%q narrationLen=%d scenes=%d)", p.ID, p.Title, len(p.Narration), len(p.Scenes))
		}
	}
}

func TestTurkishTTSPronunciationAndPhoneticAlignment(t *testing.T) {
	spoken := tubie.PrepareSpeechForTurkishTTS("Sahte APK'yı ve crack'i indirirsen, Wi-Fi'de QR veya 2FA ile phishing ve deepfake yapan hackerlar SQL ile vurur; elinde ip ile ailenin takip adımları güvende!")
	checks := []string{
		"a pe ka'yı",
		"kırek'i",
		"vayfay'de",
		"kü ar",
		"iki ef ey",
		"fişing",
		"dipfeyk",
		"hekırlar",
		"es kü el",
		"elinde ip ile ailenin takip adımları güvende!",
	}
	for _, want := range checks {
		if !strings.Contains(spoken, want) {
			t.Errorf("PrepareSpeechForTurkishTTS missing %q in:\n%s", want, spoken)
		}
	}

	crackli := tubie.PrepareSpeechForTurkishTTS("Ücretsiz crackli uygulama")
	if !strings.Contains(crackli, "kırekli") {
		t.Errorf("expected agglutinated 'crackli' -> 'kırekli', got %q", crackli)
	}

	canonical := []string{"SAHTE", "APK", "VE", "CRACK"}
	recognized := []tubie.TimedWord{
		{Word: "sahte", Start: 0.1, End: 0.4},
		{Word: "a", Start: 0.45, End: 0.55},
		{Word: "pe", Start: 0.55, End: 0.65},
		{Word: "ka", Start: 0.65, End: 0.85},
		{Word: "ve", Start: 0.90, End: 1.05},
		{Word: "kırek", Start: 1.10, End: 1.50},
	}
	aligned := tubie.AlignScriptWords(canonical, recognized, nil, 2.0)
	if len(aligned) != 4 {
		t.Fatalf("expected 4 aligned words, got %d", len(aligned))
	}
	if aligned[1].Word != "APK" || math.Abs(aligned[1].Start-0.45) > 0.05 || math.Abs(aligned[1].End-0.85) > 0.05 {
		t.Errorf("expected APK anchored to merged phonetic [0.45..0.85], got %+v", aligned[1])
	}
	if aligned[3].Word != "CRACK" || math.Abs(aligned[3].Start-1.10) > 0.05 || math.Abs(aligned[3].End-1.50) > 0.05 {
		t.Errorf("expected CRACK anchored to phonetic [1.10..1.50], got %+v", aligned[3])
	}
}
