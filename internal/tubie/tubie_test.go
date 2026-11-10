package tubie_test

import (
	"math"
	"os"
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
		if math.Abs(start-currStart) > 0.002 {
			t.Errorf("shot %d: start %.3f != expected %.3f", i, start, currStart)
		}
		if end <= start {
			t.Errorf("shot %d: end <= start", i)
		}
		currStart = end
	}
	if math.Abs(currStart-totalDur) > 0.002 {
		t.Errorf("final shot end %.3f does not match totalDur %.3f", currStart, totalDur)
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
	// Verify Instagram 2200 character limit
	if len([]rune(text)) > 2200 {
		t.Fatalf("formatted text exceeds 2200 character limit: %d characters", len([]rune(text)))
	}
}

