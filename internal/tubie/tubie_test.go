package tubie_test

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"tubie-stickman/internal/tubie"
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
	if len(presets) != 4 {
		t.Fatalf("expected 4 presets, got %d", len(presets))
	}
	for _, p := range presets {
		if p.ID == "" || p.Title == "" || p.Slug == "" || p.Narration == "" {
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
