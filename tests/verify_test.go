package verify_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repository root containing go.mod")
		}
		dir = parent
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestVerifyABPreservation(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")

	expected := map[string]string{
		"style-catalog.md":              "0c75bdb6fefab28ac6f0dc61f6fa151a0793667479e23ae2696f4964fd4ea01d",
		"prehistoric-dark-comedy.md":    "1a3a58a22b392cc087a0aa645b3f8478a3136e7e7f35d086ddb588ac091df860",
		"cozy-storytelling.md":          "a0b3f5890857582bb1e7bfa819cefcc7ffd55af72ab8997e1b286b7ad27f1419",
		"historical-cartoon.md":         "360f374c7b521353f5fbb9cb546b6f9f70a098973b99281c28d04ce28e77b01d",
		"voice-continuity.md":           "ac14fce4f8e4176815d05260de8eeea22345c0920d8e3e2b3911a380affe2ed3",
		"storyboard-template.md":        "486338e79397c73443fb390ff03883e5712834b1ddf542c3af54a8dbf6cb870b",
		"omni-flash-prompt-contract.md": "2ef4527750d10c20274f89f86cca9d6483430cef489722c4afb9893ae1d8fbe5",
		"examples.md":                   "c183e159304eb6dc12f9ad0b06f037a3f880282bc0b75bf06a2863705ddf00f4",
	}

	type refinement struct {
		old []byte
		new []byte
	}
	a5SummaryRefinements := map[string]refinement{
		"SKILL.md": {
			old: []byte("  - `Style 5 (Historical Cartoon / 历史卡通风)`: white-filled round heads, simple black oval eyes and thin black stick limbs in era-specific costume, with cel shading and full-color painterly historical settings"),
			new: []byte("  - `Style 5 (Historical Cartoon / 历史卡通风)`: white heads with connected short white necks, expressive faces, black stick limbs and compact fingerless palms, period dress and cel-shaded historical settings"),
		},
		"style-catalog.md": {
			old: []byte("| **Style 5** | Historical Cartoon (历史卡通风) | Full-color painterly historical settings, clean cel shading, era-specific wardrobe and props | History explainers, historical daily life, civilization stories and historical comedy | White filled round heads, simple black oval eyes, thick outlines and thin black stick limbs |"),
			new: []byte("| **Style 5** | Historical Cartoon (历史卡通风) | Full-color painterly historical settings, clean cel shading, era-specific wardrobe and props | History explainers, historical daily life, civilization stories and historical comedy | White filled round heads with connected short white necks, expressive individual faces, solid black oval eyes, thin black exposed limbs and small solid black mitten-shaped palms continuous with the wrists, one subtle integrated thumb lobe, no individual fingers, finger separations, white fills or gloves |"),
		},
		"omni-flash-prompt-contract.md": {
			old: []byte("Style 5 (Historical Cartoon / 历史卡通风): read `historical-cartoon.md` for its standalone prompt order and overrides. Repeat the full frozen drawing anchor, visible cast cards, approved era pack and complete `voice-continuity.md` voice block in every clip. Keep solid oval eyes and black stick limbs, with historical costume and colored painterly scenery. Do not inherit Style 1's hollow-head/no-face/no-clothes contract or generic piano, three-color, four-device and no-idle requirements. Follow its era score, physical beats and planned cuts; do not shorten approved anchors to meet an arbitrary character budget."),
			new: []byte("Style 5 (Historical Cartoon / 历史卡通风): read `historical-cartoon.md` for its standalone prompt order and overrides. Repeat the full frozen drawing anchor, visible cast cards, approved era pack and complete `voice-continuity.md` voice block in every clip. Keep connected short white necks, expressive individual faces, solid oval eyes, thin black exposed limbs and small solid black mitten-shaped palms continuous with the wrists, one subtle integrated thumb lobe, no individual fingers, finger separations, white fills or gloves, with historical costume and colored painterly scenery. Do not inherit Style 1's hollow-head/no-face/no-clothes contract or generic piano, three-color, four-device and no-idle requirements. Follow its era score, physical beats and planned cuts; do not shorten approved anchors to meet an arbitrary character budget."),
		},
	}

	preBBytes := func(name string, content []byte) []byte {
		content = bytes.ReplaceAll(content, []byte("docs/qa/reports/"), []byte("tests/reports/"))
		if ref, ok := a5SummaryRefinements[name]; ok {
			if bytes.Count(content, ref.new) != 1 {
				t.Fatalf("Missing exact A5 refinement in %s", name)
			}
			content = bytes.Replace(content, ref.new, ref.old, 1)
		}
		return content
	}

	a5ContactMittenSha := "524397d4cfbfad03a1e512eddba9883746ced30af098d30630b50e620fd5be45"

	skillMDBytes, err := os.ReadFile(filepath.Join(skill, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}
	entry := preBBytes("SKILL.md", skillMDBytes)
	marker := []byte("## A \xc2\xb7 Director\n")
	idx := bytes.Index(entry, marker)
	if idx == -1 {
		t.Fatalf("marker ## A · Director\\n not found in SKILL.md")
	}
	director := entry[idx:]
	if d := sha256Hex(director); d != "f710612be73c952fb130d891b0650388ec5ae89469cae95eda8e147eec18b73e" {
		t.Fatalf("A Director contract changed: got %s", d)
	}

	for name, expectedDigest := range expected {
		content, err := os.ReadFile(filepath.Join(skill, "references", name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		if name == "historical-cartoon.md" {
			if d := sha256Hex(content); d != a5ContactMittenSha {
				t.Fatalf("Unreviewed A5 reference change in %s: got %s, want %s", name, d, a5ContactMittenSha)
			}
		} else {
			if d := sha256Hex(preBBytes(name, content)); d != expectedDigest {
				t.Fatalf("A reference changed: %s: got %s, want %s", name, d, expectedDigest)
			}
		}
	}
}

func TestVerifyHistoricalCartoon(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")

	refBytes, err := os.ReadFile(filepath.Join(skill, "references", "historical-cartoon.md"))
	if err != nil {
		t.Fatalf("reading historical-cartoon.md: %v", err)
	}
	ref := string(refBytes)

	idx1 := strings.Index(ref, "```text\n")
	if idx1 == -1 {
		t.Fatal("missing ```text\\n in historical-cartoon.md")
	}
	sub := ref[idx1+len("```text\n"):]
	idx2 := strings.Index(sub, "```")
	if idx2 == -1 {
		t.Fatal("missing closing ``` in historical-cartoon.md")
	}
	anchor := sub[:idx2]
	if d := sha256Hex([]byte(anchor)); d != "f7f8161688c3c17a4379eaf5e4d64d7c0888f7e4b94ef59b7beb3dcafc785a8e" {
		t.Fatalf("anchor SHA-256 mismatch: got %s", d)
	}

	requiredTokens := []string{
		"https://github.com/leksautomate", "/pull/41",
		"repeat its **full description**", "two visible non-color differences",
		"no style-specific character-count cap", "voice-continuity.md",
		"default when no other voice is requested", "opening/action/ending states",
		"Backgrounds are full-color", "do not rewrite Styles 1–4",
		"black mitten-palm design with integrated thumb lobe and wrist-to-sleeve attachment",
		"not sleeves or footwear", "Grips, releases and gestures may change pose",
		"This clarification is A5-only",
		"compact black mitten palms and continuous wrist attachment during walking, grips, releases and empty-hand gestures",
		"no individual fingers, finger separations, white fills or gloves",
		"contact or overlap between the palm and the prop",
		"without spreading fingers",
		"NO INDIVIDUAL FINGERS, NO FINGER SEPARATIONS, NO WHITE HANDS OR GLOVES",
		"not frozen poses or permanently clenched fists",
		"Earlier round-tip or exposed-neck tests do not validate this revised silhouette",
		"short, sturdy white neck", "one continuous filled silhouette",
		"about one tenth of the head diameter", "roughly four head-heights tall",
		"character-specific eyebrows, nose and mouth", "18% of head diameter long and 10% wide",
		"stickman animated video", "smaller than half an eye's height",
		"On A5 selection", "front/side/back", "every clip in a supporting tool",
		"Users can knowingly choose text-only", "not evidence of reliable text-only generation",
		"rather than offering that advice only when asked", "only list specific reference files when the user actually supplies them",
		"generate their own front/side/back", "This is advice only",
		"do not supply or automatically generate reference images",
		"This connection rule is A5-only",
		"continuous white jaw–short-neck–garment connection and any beard/collar overlap",
		"A background-colored gap fails",
		"The owner reference defines the target, not cross-clip stability",
		"not overall stability passes",
	}
	for _, tok := range requiredTokens {
		if !strings.Contains(ref, tok) {
			t.Fatalf("missing required token in historical-cartoon.md: %q", tok)
		}
	}

	obsoleteTokens := []string{
		"Small white-filled, black-outlined cartoon hands attach naturally",
		"Hands are small simplified white-filled shapes",
		"small white-filled black-outlined hands connected at the wrists",
		"with minimal finger strokes or curved grips",
		"A few short finger strokes can indicate an open hand",
		"A short thin black neck line connects the base of each head to the neckline",
		"with the same thin black line treatment as the exposed limbs",
		"A short white-filled, black-outlined neck smoothly connects the base of each head to the neckline",
		"Show a clearly visible short white neck",
		"one-eighth of the head diameter",
		"must not conceal the entire white connection",
		"Do not use a black stick neck, fuse the head into the shoulders or lengthen the neck",
		"mark the neck check failed or unverified rather than passed",
		"The head joins the clothed body naturally at the neckline; no separately visible neck is required.",
		"Hands are small solid black round or oval tips",
		"NO THUMBS",
		"The base of the head may meet or overlap the garment neckline directly",
		"a short collar connection or fully concealed neck is acceptable",
		"No separate neck is exposed", "no nose, and thin black stick limbs",
		"The lower edge of the white round head physically touches",
		"about 1.5-2 wrist-stroke widths", "Keep palm length below one sixth",
		"Reference images remain optional when requested and supported",
	}
	for _, obs := range obsoleteTokens {
		if strings.Contains(ref, obs) {
			t.Fatalf("obsolete token present in historical-cartoon.md: %q", obs)
		}
	}

	for _, filename := range []string{
		"SKILL.md", "references/style-catalog.md",
		"references/storyboard-template.md", "references/omni-flash-prompt-contract.md",
	} {
		content, err := os.ReadFile(filepath.Join(skill, filename))
		if err != nil {
			t.Fatalf("reading %s: %v", filename, err)
		}
		if !strings.Contains(string(content), "historical-cartoon.md") {
			t.Fatalf("missing reference to historical-cartoon.md in %s", filename)
		}
	}

	for _, filename := range []string{
		"SKILL.md", "references/style-catalog.md", "references/omni-flash-prompt-contract.md",
	} {
		content, err := os.ReadFile(filepath.Join(skill, filename))
		if err != nil {
			t.Fatalf("reading %s: %v", filename, err)
		}
		lines := strings.Split(string(content), "\n")
		var a5Line string
		for _, line := range lines {
			if strings.Contains(line, "Style 5") {
				a5Line = line
				break
			}
		}
		if a5Line == "" {
			t.Fatalf("missing Style 5 line in %s", filename)
		}
		if filename == "SKILL.md" {
			if !strings.Contains(a5Line, "compact fingerless palms") || !strings.Contains(a5Line, "connected short white necks") {
				t.Fatalf("missing required terms in Style 5 line in SKILL.md: %s", a5Line)
			}
		} else {
			if !strings.Contains(a5Line, "small solid black mitten-shaped palms continuous with the wrists") ||
				!strings.Contains(a5Line, "one subtle integrated thumb lobe, no individual fingers, finger separations, white fills or gloves") ||
				!strings.Contains(a5Line, "connected short white necks") {
				t.Fatalf("missing required terms in Style 5 line in %s: %s", filename, a5Line)
			}
		}
		if strings.Contains(a5Line, "white-filled black-outlined cartoon hands") {
			t.Fatalf("obsolete hand description in %s: %s", filename, a5Line)
		}
	}

	baseline := map[string]string{
		"prehistoric-dark-comedy.md": "1a3a58a22b392cc087a0aa645b3f8478a3136e7e7f35d086ddb588ac091df860",
		"cozy-storytelling.md":       "a0b3f5890857582bb1e7bfa819cefcc7ffd55af72ab8997e1b286b7ad27f1419",
		"voice-continuity.md":        "24014bc599e237644c743f033a0b62c556ba4ed8689d382d08b1bcdce1967de2",
	}
	for filename, fingerprint := range baseline {
		content, err := os.ReadFile(filepath.Join(skill, "references", filename))
		if err != nil {
			t.Fatalf("reading %s: %v", filename, err)
		}
		content = bytes.ReplaceAll(content, []byte("docs/qa/reports/"), []byte("tests/reports/"))
		if filename == "voice-continuity.md" {
			heading := []byte("# Shared voice continuity \xe2\x80\x94 all visual styles\n\n")
			notice := []byte("Scope: **A \xc2\xb7 Director text-to-video prompts only**. B \xc2\xb7 Autopilot uses named TTS voices under `autopilot.md`; " +
				"do not copy this document's descriptive voice profiles into any B-style TTS request.\n\n")
			prefix := append(heading, notice...)
			if !bytes.HasPrefix(content, prefix) {
				t.Fatalf("missing exact A-only voice scope in voice-continuity.md")
			}
			content = append(heading, content[len(prefix):]...)
		}
		if d := sha256Hex(content); d != fingerprint {
			t.Fatalf("fingerprint mismatch for %s: got %s, want %s", filename, d, fingerprint)
		}
	}
}

func extractPromptBlocks(text string) []string {
	re := regexp.MustCompile("(?s)```text\n(.*?)\n```")
	matches := re.FindAllStringSubmatch(text, -1)
	var blocks []string
	for _, m := range matches {
		blocks = append(blocks, m[1])
	}
	return blocks
}

func checkVoiceVideo(prompts []string) (string, error) {
	if len(prompts) == 0 {
		return "", fmt.Errorf("no prompts")
	}
	narratorRe := regexp.MustCompile(`(?m)^One off-screen adult .* narrator throughout\..*$`)
	var voices []string
	for _, prompt := range prompts {
		matches := narratorRe.FindAllString(prompt, -1)
		if len(matches) != 1 {
			return "", fmt.Errorf("missing, abbreviated or duplicate complete narrator block (%d matches)", len(matches))
		}
		block := matches[0]
		voiceOrBaritone := "baritone"
		if strings.Contains(block, "female") {
			voiceOrBaritone = "voice"
		}
		fields := []string{
			"accent", voiceOrBaritone, "resonance",
			"raspiness", "breathiness", "delivery", "pitch variation",
			"speaking pace", "studio sound", "vocal timbre, register",
			"Emotional progression belongs to the visuals and music",
		}
		for _, f := range fields {
			if !strings.Contains(block, f) {
				return "", fmt.Errorf("incomplete profile: %s", f)
			}
		}
		if strings.Contains(block, "...") || strings.Contains(block, "[") {
			return "", fmt.Errorf("unresolved/abbreviated voice")
		}
		voices = append(voices, block)
	}
	for _, v := range voices[1:] {
		if v != voices[0] {
			return "", fmt.Errorf("voice profiles drift within this video")
		}
	}
	return voices[0], nil
}

func TestVerifyVoiceContinuity(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")
	refs := filepath.Join(skill, "references")

	voiceDocBytes, err := os.ReadFile(filepath.Join(refs, "voice-continuity.md"))
	if err != nil {
		t.Fatalf("reading voice-continuity.md: %v", err)
	}
	voiceDoc := string(voiceDocBytes)
	voiceBlocks := extractPromptBlocks(voiceDoc)
	if len(voiceBlocks) == 0 {
		t.Fatal("no prompt blocks in voice-continuity.md")
	}
	male := voiceBlocks[0]

	v, err := checkVoiceVideo([]string{male, male, male})
	if err != nil || v != male {
		t.Fatalf("self check failed: %v", err)
	}

	mutations := []string{
		"Same narrator as clip 1.",
		strings.Replace(male, "medium-low pitch", "high pitch", 1),
	}
	for _, mut := range mutations {
		if _, err := checkVoiceVideo([]string{male, mut, male}); err == nil {
			t.Fatalf("Voice drift/shortcut was not rejected for mutation: %s", mut)
		}
	}

	examplesBytes, err := os.ReadFile(filepath.Join(refs, "examples.md"))
	if err != nil {
		t.Fatalf("reading examples.md: %v", err)
	}
	examples := string(examplesBytes)
	examplePrompts := extractPromptBlocks(examples)
	if len(examplePrompts) != 9 {
		t.Fatalf("expected 9 example prompts, got %d", len(examplePrompts))
	}

	female, err := checkVoiceVideo(examplePrompts[:6])
	if err != nil {
		t.Fatalf("female prompt check failed: %v", err)
	}
	if !strings.Contains(female, "adult female") || strings.Contains(female, "baritone") {
		t.Fatalf("female prompt content unexpected: %s", female)
	}

	maleEx, err := checkVoiceVideo(examplePrompts[6:])
	if err != nil {
		t.Fatalf("male prompt check failed: %v", err)
	}
	if maleEx != male {
		t.Fatalf("male prompt did not match voice-continuity male block")
	}
	if female == male {
		t.Fatal("Different videos cannot choose different voices")
	}

	for _, filename := range []string{
		"SKILL.md", "references/style-catalog.md", "references/storyboard-template.md",
		"references/omni-flash-prompt-contract.md", "references/prehistoric-dark-comedy.md",
		"references/cozy-storytelling.md",
	} {
		c, err := os.ReadFile(filepath.Join(skill, filename))
		if err != nil {
			t.Fatalf("reading %s: %v", filename, err)
		}
		if !strings.Contains(string(c), "voice-continuity.md") {
			t.Fatalf("missing shared route: %s", filename)
		}
	}

	for _, required := range []string{
		"Style 1 Light/Dark, Style 2A, Style 2B and Style 3",
		"Style 4 Cozy Storytelling",
		"multi-speaker", "explicitly silent video",
		"Other voice profiles and styles have not had separate provider-backed audio comparisons",
	} {
		if !strings.Contains(voiceDoc, required) {
			t.Fatalf("missing required text in voice-continuity.md: %s", required)
		}
	}

	prefixes := []string{
		"Character:", "Environment:", "First frame:", "[0–3s]", "[3–7s]", "[7–10s]",
		"Audio-only dialogue", "Do not generate", "Negative constraints:",
	}
	var nonVoiceLines []string
	for _, prompt := range examplePrompts {
		for _, line := range strings.Split(prompt, "\n") {
			for _, pfx := range prefixes {
				if strings.HasPrefix(line, pfx) {
					nonVoiceLines = append(nonVoiceLines, line)
					break
				}
			}
		}
	}
	nonVoice := strings.Join(nonVoiceLines, "\n")
	if d := sha256Hex([]byte(nonVoice)); d != "bbd7574179ba98c8cc6ceadadecfa0b5bc2462c312e57dc810616d114fc71403" {
		t.Fatalf("Non-voice example content changed: got %s", d)
	}
}

func TestVerifyFirstUseGuidance(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")

	entryBytes, err := os.ReadFile(filepath.Join(skill, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}
	entry := string(entryBytes)

	guidePath := filepath.Join(skill, "references", "getting-started.md")
	guideBytes, err := os.ReadFile(guidePath)
	if err != nil {
		t.Fatalf("reading getting-started.md: %v", err)
	}
	guideText := string(guideBytes)

	if !strings.Contains(entry, "(references/getting-started.md)") {
		t.Fatal("SKILL.md missing (references/getting-started.md)")
	}

	mdLinkRe := regexp.MustCompile(`\]\(([^)]+\.md)(?:#[^)]*)?\)`)
	for _, src := range []struct {
		name string
		text string
		dir  string
	}{
		{"SKILL.md", entry, skill},
		{"getting-started.md", guideText, filepath.Dir(guidePath)},
	} {
		matches := mdLinkRe.FindAllStringSubmatch(src.text, -1)
		for _, m := range matches {
			target := m[1]
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
				continue
			}
			full := filepath.Join(src.dir, target)
			if fi, err := os.Stat(full); err != nil || fi.IsDir() {
				t.Fatalf("broken reference in %s: %s -> %s", src.name, target, full)
			}
		}
	}

	marker := "## A · Director\n"
	idx := strings.Index(entry, marker)
	if idx == -1 {
		t.Fatal("missing ## A · Director in SKILL.md")
	}
	legacy := entry[idx:]
	originalA5 := "  - `Style 5 (Historical Cartoon / 历史卡通风)`: white-filled round heads, simple black oval eyes and thin black stick limbs in era-specific costume, with cel shading and full-color painterly historical settings"
	revisedA5 := "  - `Style 5 (Historical Cartoon / 历史卡通风)`: white heads with connected short white necks, expressive faces, black stick limbs and compact fingerless palms, period dress and cel-shaded historical settings"
	if strings.Count(legacy, revisedA5) != 1 {
		t.Fatal("Missing exact owner-directed A5 baseline summary in SKILL.md legacy")
	}
	legacy = strings.Replace(legacy, revisedA5, originalA5, 1)
	if d := sha256Hex([]byte(legacy)); d != "f710612be73c952fb130d891b0650388ec5ae89469cae95eda8e147eec18b73e" {
		t.Fatalf("legacy director digest mismatch: got %s", d)
	}

	if len(entryBytes) > 15000 {
		t.Fatalf("SKILL.md exceeds 15000 bytes: %d", len(entryBytes))
	}

	boundaries := filepath.Join(skill, "references", "mode-execution-boundaries.md")
	if fi, err := os.Stat(boundaries); err != nil || fi.IsDir() {
		t.Fatal("mode-execution-boundaries.md missing")
	}
	if !strings.Contains(entry, "(references/mode-execution-boundaries.md)") {
		t.Fatal("SKILL.md missing link to mode-execution-boundaries.md")
	}
	if !strings.Contains(entry, "Before planning") {
		t.Fatal("SKILL.md missing 'Before planning'")
	}

	for _, token := range []string{
		"Published B choices are B1, B2, B3, B4 and B5",
		"scripts/character-explainer.mjs",
		"B1–B4 use the bundled manifest runtime",
	} {
		if !strings.Contains(guideText, token) {
			t.Fatalf("missing required text in getting-started.md: %s", token)
		}
	}
	for _, forbidden := range []string{
		"not an accepted `style: B4` CLI value",
		"B4 and B5 remain local development only",
		"Only B1 is currently available",
	} {
		if strings.Contains(guideText, forbidden) {
			t.Fatalf("forbidden text present in getting-started.md: %s", forbidden)
		}
	}

	// Frontmatter
	fmParts := strings.SplitN(entry, "---\n", 3)
	if len(fmParts) < 3 {
		t.Fatal("SKILL.md missing frontmatter")
	}
	fm := fmParts[1]
	nameRe := regexp.MustCompile(`(?m)^name:\s*(.+)$`)
	descRe := regexp.MustCompile(`(?m)^description:\s*(.+)$`)
	nameMatch := nameRe.FindStringSubmatch(fm)
	descMatch := descRe.FindStringSubmatch(fm)
	if len(nameMatch) < 2 || len(descMatch) < 2 {
		t.Fatal("missing name or description in SKILL.md frontmatter")
	}
	skillName := strings.TrimSpace(nameMatch[1])
	if skillName != "directing-stickman-videos" {
		t.Fatalf("unexpected skill name: %s", skillName)
	}
	var descVal string
	if err := json.Unmarshal([]byte(strings.TrimSpace(descMatch[1])), &descVal); err != nil || descVal == "" {
		t.Fatalf("invalid description in SKILL.md frontmatter: %v", descMatch[1])
	}

	// agents/openai.yaml
	openaiYAMLBytes, err := os.ReadFile(filepath.Join(skill, "agents", "openai.yaml"))
	if err != nil {
		t.Fatalf("reading openai.yaml: %v", err)
	}
	openaiYAML := string(openaiYAMLBytes)
	dpRe := regexp.MustCompile(`(?m)^\s*default_prompt:\s*(.+)$`)
	dnRe := regexp.MustCompile(`(?m)^\s*display_name:\s*(.+)$`)
	dpMatch := dpRe.FindStringSubmatch(openaiYAML)
	dnMatch := dnRe.FindStringSubmatch(openaiYAML)
	if len(dpMatch) < 2 || len(dnMatch) < 2 {
		t.Fatal("missing default_prompt or display_name in openai.yaml")
	}
	var defaultPrompt, displayName string
	if err := json.Unmarshal([]byte(strings.TrimSpace(dpMatch[1])), &defaultPrompt); err != nil {
		t.Fatalf("unmarshal default_prompt: %v", err)
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(dnMatch[1])), &displayName); err != nil {
		t.Fatalf("unmarshal display_name: %v", err)
	}
	if !strings.Contains(defaultPrompt, "$"+skillName) {
		t.Fatalf("default_prompt does not contain $%s", skillName)
	}
	if displayName != "Stickman Video Director" {
		t.Fatalf("displayName is %q, expected 'Stickman Video Director'", displayName)
	}

	// Single SKILL.md in repo
	var skillMDs []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
		} else if info.Name() == "SKILL.md" {
			skillMDs = append(skillMDs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking repo: %v", err)
	}
	expectedSkillPath := filepath.Join(skill, "SKILL.md")
	if len(skillMDs) != 1 || skillMDs[0] != expectedSkillPath {
		t.Fatalf("expected single SKILL.md at %s, found %v", expectedSkillPath, skillMDs)
	}

	// package.json scripts check
	pkgBytes, err := os.ReadFile(filepath.Join(skill, "package.json"))
	if err != nil {
		t.Fatalf("reading package.json: %v", err)
	}
	var pkg map[string]any
	if err := json.Unmarshal(pkgBytes, &pkg); err != nil {
		t.Fatalf("unmarshal package.json: %v", err)
	}
	scripts, _ := pkg["scripts"].(map[string]any)
	for _, hook := range []string{"preinstall", "install", "postinstall"} {
		if _, exists := scripts[hook]; exists {
			t.Fatalf("forbidden install hook in package.json: %s", hook)
		}
	}

	// Public catalogs
	publicLinkRe := regexp.MustCompile(`\]\((https://[^)]+)\)`)
	publicMatches := publicLinkRe.FindAllStringSubmatch(guideText, -1)
	publicLinks := make(map[string]bool)
	for _, m := range publicMatches {
		publicLinks[m[1]] = true
	}
	for _, locale := range []string{"en", "tr"} {
		catalogPath := filepath.Join(root, "docs", "catalog", fmt.Sprintf("README.%s.md", locale))
		catBytes, err := os.ReadFile(catalogPath)
		if err != nil {
			t.Fatalf("reading %s: %v", catalogPath, err)
		}
		catText := string(catBytes)
		for _, anchor := range []string{"a-styles", "b-styles"} {
			u1 := fmt.Sprintf("https://github.com/egnake/Tubie-Stickman/blob/main/docs/catalog/README.%s.md#%s", locale, anchor)
			u2 := fmt.Sprintf("https://github.com/egnake/Tubie-Stickman/blob/main/docs/catalog/README.%s.md", locale)
			if !publicLinks[u1] && !publicLinks[u2] {
				t.Fatalf("missing public link for catalog %s", u1)
			}
			tag := fmt.Sprintf(`<a id="%s"></a>`, anchor)
			if !strings.Contains(catText, tag) {
				t.Fatalf("missing anchor tag %s in %s", tag, catalogPath)
			}
		}
	}
	scenarioPath := filepath.Join(root, "docs", "qa", "scenarios", "first-use-guidance.md")
	if fi, err := os.Stat(scenarioPath); err != nil || fi.IsDir() {
		t.Fatal("missing docs/qa/scenarios/first-use-guidance.md")
	}

	// Chinese welcome
	wParts := strings.Split(guideText, "### 欢迎使用火柴人视频 Skill\n")
	if len(wParts) < 2 {
		t.Fatal("missing Chinese welcome header")
	}
	welcome := strings.Split(wParts[1], "### Adapt instead of replaying the README")[0]
	if !strings.Contains(welcome, "#### A · 导演模式：") || !strings.Contains(welcome, "#### B · 全自动模式：") {
		t.Fatal("missing mode headings in welcome")
	}
	forbiddenRe := regexp.MustCompile(`\b(?:Director|Autopilot)\b|自动成片`)
	if forbiddenRe.MatchString(welcome) {
		t.Fatal("forbidden routing labels in Chinese welcome")
	}
	if !strings.Contains(welcome, "A 模式的旁白和声音设计沿用原有规则") {
		t.Fatal("missing narration rule statement in welcome")
	}
	if strings.Contains(welcome, "稳定男声默认") {
		t.Fatal("misstated voice selection in welcome")
	}

	// Optional A references
	refHeading := "Optional A character references"
	if !strings.Contains(entry, refHeading) {
		t.Fatal("missing refHeading in entry")
	}
	aRefParts := strings.Split(guideText, fmt.Sprintf("### %s\n", refHeading))
	if len(aRefParts) < 2 {
		t.Fatal("missing refHeading in guideText")
	}
	aRef := strings.Split(aRefParts[1], "## B · Autopilot handoffs")[0]
	for _, req := range []string{
		"not a new setup requirement", "front, side and back views",
		"Reuse suitable user-approved references", "every generated clip",
		"when the chosen tool supports reference images", "not a consistency guarantee",
		"not replace, shorten or globally rewrite", "fixed character anchor",
	} {
		if !strings.Contains(aRef, req) {
			t.Fatalf("missing requirement in aRef: %s", req)
		}
	}
	if !strings.Contains(welcome, "按需确认人物三视图") || !strings.Contains(welcome, "三视图（正面、侧面、背面）") {
		t.Fatal("missing 3-view text in welcome")
	}

	// A5 proactive character guidance
	if !strings.Contains(entry, "A5 proactive character guidance") {
		t.Fatal("missing A5 proactive character guidance in entry")
	}
	a5RefParts := strings.Split(guideText, "### A5 proactive character guidance\n")
	if len(a5RefParts) < 2 {
		t.Fatal("missing A5 proactive guidance in guideText")
	}
	a5Ref := strings.Split(a5RefParts[1], "## B · Autopilot handoffs")[0]
	for _, req := range []string{
		"before planning", "even if the user has not asked", "正面、侧面、背面",
		"每段视频", "完整分镜提示词一起提交", "not a new mandatory setup",
		"prefers text-only", "do not repeatedly ask", "Reuse suitable approved",
		"additional image input", "not a replacement", "first-frame image",
		"not claim guaranteed identity", "missing hair and disconnected heads",
		"建议你先自行生成", "advice only", "Do not bundle or supply reference images",
		"automatically generate them", "only when the user explicitly requests",
		"name specific files only when actually supplied",
	} {
		if !strings.Contains(a5Ref, req) {
			t.Fatalf("missing requirement in a5Ref: %s", req)
		}
	}

	// Defaults precedence
	autopilotBytes, err := os.ReadFile(filepath.Join(skill, "references", "autopilot.md"))
	if err != nil {
		t.Fatalf("reading autopilot.md: %v", err)
	}
	autopilotText := string(autopilotBytes)
	b1Bytes, err := os.ReadFile(filepath.Join(skill, "references", "b1-minimal-handdrawn.md"))
	if err != nil {
		t.Fatalf("reading b1-minimal-handdrawn.md: %v", err)
	}
	b1Text := string(b1Bytes)

	priority := "explicit user choice → selected B-style default → B common fallback"
	if !strings.Contains(autopilotText, priority) || !strings.Contains(guideText, priority) {
		t.Fatal("priority string missing from autopilot or guide")
	}
	if !strings.Contains(entry, "visuals, captions, voice, pacing and SFX") {
		t.Fatal("missing entry settings phrase")
	}
	if !strings.Contains(autopilotText, "caption font/size/placement") {
		t.Fatal("missing caption styling in autopilot")
	}
	if !strings.Contains(b1Text, "B1 defaults, not requirements for other B styles") {
		t.Fatal("missing B1 defaults phrase")
	}
	if !strings.Contains(b1Text, "Google Gemini **Bodi**") || !strings.Contains(b1Text, "original **Sadaltager** recording") {
		t.Fatal("missing Bodi / Sadaltager in b1Text")
	}
	if !strings.Contains(b1Text, "without added style/direction instructions") {
		t.Fatal("missing instruction phrase in b1Text")
	}
}

func TestVerifyBCommonQuality(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")
	refs := filepath.Join(skill, "references")

	entryBytes, err := os.ReadFile(filepath.Join(skill, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}
	entry := string(entryBytes)
	preDirector := strings.Split(entry, "## A · Director")[0]
	if !strings.Contains(preDirector, "(references/autopilot.md)") {
		t.Fatal("pre-Director missing (references/autopilot.md)")
	}

	autopilotBytes, err := os.ReadFile(filepath.Join(refs, "autopilot.md"))
	if err != nil {
		t.Fatalf("reading autopilot.md: %v", err)
	}
	autopilot := string(autopilotBytes)
	if !strings.Contains(autopilot, "(autopilot-production-qa.md)") ||
		!strings.Contains(autopilot, "(autopilot-production-qa.md#whole-film-delivery-gate)") ||
		!strings.Contains(autopilot, "before shot planning, sample review and full-film delivery") ||
		!strings.Contains(autopilot, "Style defaults customize expression, not waive this common quality floor") {
		t.Fatal("missing required routing/boundary in autopilot.md")
	}

	commonBytes, err := os.ReadFile(filepath.Join(refs, "autopilot-production-qa.md"))
	if err != nil {
		t.Fatalf("reading autopilot-production-qa.md: %v", err)
	}
	common := string(commonBytes)
	if !strings.Contains(common, "**every B style**") || !strings.Contains(common, "A does not inherit it") {
		t.Fatal("missing every B style / A does not inherit it in common QA")
	}

	for _, heading := range []string{
		"Plan from meaning and actual narration",
		"First 30 seconds: picture, sound and copy together",
		"Meaningful motion, not filler",
		"Clear screenshots and content-matched graphics",
		"Natural captions and semantic sound",
		"Whole-film delivery gate",
	} {
		if !strings.Contains(common, fmt.Sprintf("## %s", heading)) {
			t.Fatalf("missing common gate heading: ## %s", heading)
		}
	}

	for _, boundary := range []string{
		"propose any needed hook rewrite for approval",
		"approximately 20-second sample gate remains unchanged",
		"principal-picture inactivity longer than 2.5 seconds",
		"not a maximum shot length",
		"Subtitle turnover, whole-frame zoom/pan/fade",
		"Do not repeat compositions to fill time",
		"final encoded frame",
		"actual entity/action",
		"Review every cue boundary against meaning and actual speech pauses",
		"semantic subtitles with at most one visible line at a time",
		"Preserve all spoken words and order",
		"noun phrases, verbs with necessary objects",
		"numbers/units and phrasal verbs",
		"rather than a second line or mechanically smaller/squeezed type",
		"neighboring cues for broken meaning",
		"each cue starts at its own first aligned spoken word",
		"same detail standard throughout the entire film",
		"excluding the subtitle band",
		"validator does not automatically enforce these creative gates",
		"no new paid calls",
	} {
		if !strings.Contains(common, boundary) {
			t.Fatalf("missing boundary statement in common QA: %s", boundary)
		}
	}

	for _, b1Setting := range []string{"Bodi", "Sadaltager", "34px", "#FDFDFC", "60–70%"} {
		if strings.Contains(common, b1Setting) {
			t.Fatalf("B1-specific setting leaked into common QA: %s", b1Setting)
		}
	}

	b1Bytes, err := os.ReadFile(filepath.Join(refs, "b1-minimal-handdrawn.md"))
	if err != nil {
		t.Fatalf("reading b1-minimal-handdrawn.md: %v", err)
	}
	b1 := string(b1Bytes)
	supplementBytes, err := os.ReadFile(filepath.Join(refs, "b1-production-qa.md"))
	if err != nil {
		t.Fatalf("reading b1-production-qa.md: %v", err)
	}
	supplement := string(supplementBytes)

	if !strings.Contains(b1, "(autopilot-production-qa.md)") || !strings.Contains(supplement, "(autopilot-production-qa.md)") {
		t.Fatal("missing autopilot-production-qa.md in b1 or supplement")
	}
	if !strings.Contains(supplement, "B1-only visual/audio defaults and accepted examples") ||
		!strings.Contains(supplement, "common quality gates still apply to every B production") {
		t.Fatal("missing quality gate text in supplement")
	}

	manifestBytes, err := os.ReadFile(filepath.Join(skill, "assets", "b1", "approved-production-standard.json"))
	if err != nil {
		t.Fatalf("reading approved-production-standard.json: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("unmarshal approved-production-standard.json: %v", err)
	}
	if manifest["common_checklist"] != "../../references/autopilot-production-qa.md" {
		t.Fatalf("unexpected common_checklist: %v", manifest["common_checklist"])
	}
	scope, _ := manifest["scope"].(map[string]any)
	notInheritedScope, _ := scope["not_inherited_scope"].(string)
	if !strings.Contains(notInheritedScope, "not the independently shared B") {
		t.Fatalf("unexpected not_inherited_scope: %s", notInheritedScope)
	}

	runtimeBytes, err := os.ReadFile(filepath.Join(refs, "autopilot-runtime.md"))
	if err != nil {
		t.Fatalf("reading autopilot-runtime.md: %v", err)
	}
	runtime := string(runtimeBytes)
	b2Bytes, err := os.ReadFile(filepath.Join(refs, "b2-colorful-handdrawn.md"))
	if err != nil {
		t.Fatalf("reading b2-colorful-handdrawn.md: %v", err)
	}
	b2 := string(b2Bytes)

	if !strings.Contains(runtime, "accepts `style: B1`, `B2`, `B3` and `B4`") ||
		!strings.Contains(runtime, "## B4 production tracks") ||
		!strings.Contains(autopilot, "(b2-colorful-handdrawn.md)") ||
		!strings.Contains(b2, "independent local illustration-and-micro-animation style") ||
		!strings.Contains(b2, "Fola") || !strings.Contains(b2, "ivory paper label") ||
		!strings.Contains(runtime, "B1 `pose` atlas is rejected in B2") ||
		!strings.Contains(runtime, "`style: B3`") || !strings.Contains(runtime, "## B3 layered artwork") {
		t.Fatal("missing runtime / b2 styling declarations")
	}

	b3Bytes, err := os.ReadFile(filepath.Join(refs, "b3-television.md"))
	if err != nil {
		t.Fatalf("reading b3-television.md: %v", err)
	}
	b3 := string(b3Bytes)
	if !strings.Contains(autopilot, "(b3-television.md)") || !strings.Contains(entry, "(references/b3-television.md)") {
		t.Fatal("missing b3-television.md in autopilot or entry")
	}
	if !strings.Contains(strings.ToLower(b3), "unit tests") || !strings.Contains(b3, "not animation quality or acceptance") {
		t.Fatal("missing unit test / quality statement in b3")
	}
	if strings.Contains(b3, "currently accepts B1/B2 only") {
		t.Fatal("b3 contains obsolete restriction: currently accepts B1/B2 only")
	}

	// Portable links check
	linkRe := regexp.MustCompile(`\]\(([^)]+)\)`)
	for _, docPath := range []string{
		filepath.Join(refs, "autopilot-production-qa.md"),
		filepath.Join(refs, "autopilot.md"),
		filepath.Join(refs, "b1-production-qa.md"),
		filepath.Join(refs, "b2-colorful-handdrawn.md"),
		filepath.Join(refs, "b3-television.md"),
	} {
		cBytes, err := os.ReadFile(docPath)
		if err != nil {
			t.Fatalf("reading %s: %v", docPath, err)
		}
		cText := string(cBytes)
		if strings.Contains(cText, "/Users/") {
			t.Fatalf("found /Users/ in %s", docPath)
		}
		matches := linkRe.FindAllStringSubmatch(cText, -1)
		for _, m := range matches {
			link := m[1]
			if strings.Contains(link, "://") {
				continue
			}
			targetPath := strings.Split(link, "#")[0]
			full := filepath.Clean(filepath.Join(filepath.Dir(docPath), targetPath))
			rel, err := filepath.Rel(skill, full)
			if err != nil || strings.HasPrefix(rel, "..") {
				t.Fatalf("link escapes skill directory in %s: %s -> %s", docPath, link, full)
			}
			if fi, err := os.Stat(full); err != nil || fi.IsDir() {
				t.Fatalf("broken relative link in %s: %s -> %s", docPath, link, full)
			}
		}
	}

	casesBytes, err := os.ReadFile(filepath.Join(root, "docs", "qa", "scenarios", "b-common-production-quality.md"))
	if err != nil {
		t.Fatalf("reading b-common-production-quality.md: %v", err)
	}
	cases := string(casesBytes)
	for _, caseID := range []string{
		"B-COMMON-01", "B-COMMON-02", "B-COMMON-03", "B-COMMON-04",
		"B-COMMON-05", "B-COMMON-06", "B-COMMON-07", "B-COMMON-08",
	} {
		if !strings.Contains(cases, caseID) {
			t.Fatalf("missing case %s in b-common-production-quality.md", caseID)
		}
	}
}

func TestVerifyB1Baseline(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")
	assets := filepath.Join(skill, "assets", "b1")

	manifestBytes, err := os.ReadFile(filepath.Join(assets, "approved-baseline.json"))
	if err != nil {
		t.Fatalf("reading approved-baseline.json: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("unmarshal approved-baseline.json: %v", err)
	}

	if manifest["schema"] != "b1-reference-evidence-v1" {
		t.Fatalf("unexpected schema: %v", manifest["schema"])
	}
	if manifest["style"] != "B1" {
		t.Fatalf("unexpected style: %v", manifest["style"])
	}
	if manifest["baseline_id"] != "b1-2026-10-23-ryan-v4" {
		t.Fatalf("unexpected baseline_id: %v", manifest["baseline_id"])
	}

	evidence, _ := manifest["evidence"].([]any)
	kinds := make(map[string]bool)
	for _, e := range evidence {
		eMap := e.(map[string]any)
		kinds[eMap["kind"].(string)] = true
	}
	for _, expectedKind := range []string{
		"video_with_narration_captions_and_sfx",
		"contact_sheet",
		"encoded_motion_detail_sheet",
	} {
		if !kinds[expectedKind] {
			t.Fatalf("missing evidence kind: %s", expectedKind)
		}
	}

	inSkill := func(rel string, base string) string {
		if filepath.IsAbs(rel) {
			t.Fatalf("reference must be portable: %s", rel)
		}
		target := filepath.Clean(filepath.Join(base, rel))
		relToSkill, err := filepath.Rel(skill, target)
		if err != nil || strings.HasPrefix(relToSkill, "..") {
			t.Fatalf("reference escapes skill: %s -> %s", rel, target)
		}
		if fi, err := os.Stat(target); err != nil || fi.IsDir() {
			t.Fatalf("missing reference file: %s", target)
		}
		return target
	}

	prodStandard := inSkill(manifest["production_standard"].(string), assets)

	for _, e := range evidence {
		eMap := e.(map[string]any)
		path := inSkill(eMap["path"].(string), assets)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading evidence file %s: %v", path, err)
		}
		if d := sha256Hex(b); d != eMap["sha256"] {
			t.Fatalf("SHA-256 mismatch for evidence %s: got %s, want %s", path, d, eMap["sha256"])
		}
	}

	docBytes, err := os.ReadFile(prodStandard)
	if err != nil {
		t.Fatalf("reading production standard: %v", err)
	}
	doc := string(docBytes)
	linkRe := regexp.MustCompile(`\]\(([^)]+)\)`)
	for _, m := range linkRe.FindAllStringSubmatch(doc, -1) {
		link := m[1]
		if !strings.Contains(link, "://") {
			inSkill(strings.Split(link, "#")[0], filepath.Dir(prodStandard))
		}
	}

	entryBytes, err := os.ReadFile(filepath.Join(skill, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}
	if !strings.Contains(string(entryBytes), "references/b1-minimal-handdrawn.md") {
		t.Fatal("SKILL.md missing references/b1-minimal-handdrawn.md")
	}

	np, _ := manifest["narration_provenance"].(map[string]any)
	if np["voice"] != "Bodi" || np["speed"] != "native_unstretched" {
		t.Fatalf("unexpected narration provenance: %v", np)
	}
	scope, _ := manifest["scope"].(map[string]any)
	showcase, _ := scope["public_showcase"].(string)
	if !strings.Contains(showcase, "Ryan V4") || strings.Contains(showcase, "not GitHub publication") {
		t.Fatalf("unexpected public showcase: %s", showcase)
	}
	if strings.Contains(doc, "public preview remains V15") {
		t.Fatal("doc contains obsolete preview note")
	}

	licBytes, err := os.ReadFile(filepath.Join(assets, "LUCIDE-LICENSE.txt"))
	if err != nil {
		t.Fatalf("reading LUCIDE-LICENSE.txt: %v", err)
	}
	licText := string(licBytes)
	if !strings.Contains(licText, "Lucide Contributors 2022") || !strings.Contains(licText, "Permission to use, copy, modify") {
		t.Fatal("invalid Lucide license text")
	}

	attribBytes, err := os.ReadFile(filepath.Join(assets, "ATTRIBUTION.md"))
	if err != nil {
		t.Fatalf("reading ATTRIBUTION.md: %v", err)
	}
	if !strings.Contains(string(attribBytes), "LUCIDE-LICENSE.txt") {
		t.Fatal("ATTRIBUTION.md does not reference LUCIDE-LICENSE.txt")
	}
}

func TestVerifyB1ProductionStandard(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")
	assets := filepath.Join(skill, "assets", "b1")
	manifestPath := filepath.Join(assets, "approved-production-standard.json")

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("reading %s: %v", manifestPath, err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("unmarshal %s: %v", manifestPath, err)
	}

	if manifest["schema"] != "b1-production-standard-v1" {
		t.Fatalf("unexpected schema: %v", manifest["schema"])
	}
	if manifest["style"] != "B1" {
		t.Fatalf("unexpected style: %v", manifest["style"])
	}
	if manifest["accepted_on"] != "2026-10-24" {
		t.Fatalf("unexpected accepted_on: %v", manifest["accepted_on"])
	}

	scope, _ := manifest["scope"].(map[string]any)
	appliesTo, _ := scope["applies_to"].(string)
	if !strings.HasPrefix(appliesTo, "B1 only") {
		t.Fatalf("unexpected applies_to: %s", appliesTo)
	}
	notInheritedBy, _ := scope["not_inherited_by"].([]any)
	notInheritedSet := make(map[string]bool)
	for _, n := range notInheritedBy {
		notInheritedSet[n.(string)] = true
	}
	for _, expectedItem := range []string{"A1", "A2", "A3", "A4", "A5", "other B styles"} {
		if !notInheritedSet[expectedItem] {
			t.Fatalf("missing not_inherited_by item: %s", expectedItem)
		}
	}

	accSource, _ := manifest["accepted_source"].(map[string]any)
	if accSource["sha256"] != "3739b0b1d8aef3678030cb4ac0c1d5ee0f36431f701a9ce0aad1adafb7f2e864" {
		t.Fatalf("unexpected accepted_source sha256: %v", accSource["sha256"])
	}
	if fmt.Sprintf("%v", accSource["frames"]) != "5817" {
		t.Fatalf("unexpected accepted_source frames: %v", accSource["frames"])
	}
	fileName, _ := accSource["file_name"].(string)
	if _, err := os.Stat(filepath.Join(assets, fileName)); err == nil {
		t.Fatalf("accepted_source file_name should not exist in assets: %s", fileName)
	}

	local := func(rel string, base string) string {
		if filepath.IsAbs(rel) {
			t.Fatalf("link must be portable: %s", rel)
		}
		target := filepath.Clean(filepath.Join(base, rel))
		relToSkill, err := filepath.Rel(skill, target)
		if err != nil || strings.HasPrefix(relToSkill, "..") {
			t.Fatalf("link escapes skill: %s -> %s", rel, target)
		}
		if fi, err := os.Stat(target); err != nil || fi.IsDir() {
			t.Fatalf("missing required file: %s", rel)
		}
		return target
	}

	standard := local(manifest["production_standard"].(string), assets)
	checklist := local(manifest["required_checklist"].(string), assets)
	commonChecklist := local(manifest["common_checklist"].(string), assets)
	if filepath.Base(commonChecklist) != "autopilot-production-qa.md" {
		t.Fatalf("unexpected common checklist base name: %s", commonChecklist)
	}
	notInheritedScope, _ := scope["not_inherited_scope"].(string)
	if !strings.Contains(notInheritedScope, "not the independently shared B") {
		t.Fatalf("unexpected not_inherited_scope: %s", notInheritedScope)
	}

	baseRef := local(manifest["base_reference"].(string), assets)
	baseRefBytes, err := os.ReadFile(baseRef)
	if err != nil {
		t.Fatalf("reading base reference: %v", err)
	}
	var baseRefMap map[string]any
	if err := json.Unmarshal(baseRefBytes, &baseRefMap); err != nil {
		t.Fatalf("unmarshal base reference: %v", err)
	}
	if baseRefMap["baseline_id"] != "b1-2026-10-23-ryan-v4" {
		t.Fatalf("unexpected baseline_id in base reference: %v", baseRefMap["baseline_id"])
	}

	evidence, _ := manifest["evidence"].([]any)
	if len(evidence) != 1 {
		t.Fatalf("expected 1 evidence item, got %d", len(evidence))
	}
	eItem := evidence[0].(map[string]any)
	if eItem["kind"] != "sampled_encoded_contact_sheet" {
		t.Fatalf("unexpected evidence kind: %v", eItem["kind"])
	}
	targetPath := local(eItem["path"].(string), assets)
	tBytes, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("reading evidence file: %v", err)
	}
	if d := sha256Hex(tBytes); d != eItem["sha256"] {
		t.Fatalf("evidence sha256 mismatch: got %s, want %s", d, eItem["sha256"])
	}

	for _, docPath := range []string{standard, checklist} {
		docBytes, err := os.ReadFile(docPath)
		if err != nil {
			t.Fatalf("reading doc %s: %v", docPath, err)
		}
		docText := string(docBytes)
		if strings.Contains(docText, "/Users/") || strings.Contains(docText, "/var/folders/") {
			t.Fatalf("absolute path found in %s", docPath)
		}
		linkRe := regexp.MustCompile(`\]\(([^)]+)\)`)
		for _, m := range linkRe.FindAllStringSubmatch(docText, -1) {
			link := m[1]
			if !strings.Contains(link, "://") {
				local(strings.Split(link, "#")[0], filepath.Dir(docPath))
			}
		}
	}

	stdBytes, _ := os.ReadFile(standard)
	chkBytes, _ := os.ReadFile(checklist)
	if !strings.Contains(string(stdBytes), "b1-production-qa.md") {
		t.Fatal("checklist must be reachable from loaded B1 reference")
	}
	if !strings.Contains(string(stdBytes), filepath.Base(commonChecklist)) || !strings.Contains(string(chkBytes), filepath.Base(commonChecklist)) {
		t.Fatal("B1 must inherit B common quality separately from its style defaults")
	}

	// Single SKILL.md in skill directory
	var skillMDs []string
	filepath.Walk(skill, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == "SKILL.md" {
			skillMDs = append(skillMDs, path)
		}
		return nil
	})
	if len(skillMDs) != 1 {
		t.Fatalf("expected 1 SKILL.md in skill dir, got %v", skillMDs)
	}

	if fi, err := os.Stat(filepath.Join(root, "docs", "qa", "scenarios", "b1-production-quality.md")); err != nil || fi.IsDir() {
		t.Fatal("missing docs/qa/scenarios/b1-production-quality.md")
	}
	if strings.Contains(string(manifestBytes), "/Users/") {
		t.Fatal("found /Users/ in manifest JSON")
	}
}

func TestVerifyClassicCharacter(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")

	catBytes, err := os.ReadFile(filepath.Join(skill, "references", "style-catalog.md"))
	if err != nil {
		t.Fatalf("reading style-catalog.md: %v", err)
	}
	catalog := string(catBytes)

	contractBytes, err := os.ReadFile(filepath.Join(skill, "references", "omni-flash-prompt-contract.md"))
	if err != nil {
		t.Fatalf("reading omni-flash-prompt-contract.md: %v", err)
	}
	contract := string(contractBytes)

	skillBytes, err := os.ReadFile(filepath.Join(skill, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}
	skillText := string(skillBytes)

	p1 := strings.Split(catalog, "## 2. Style 1: Classic Minimalist")
	if len(p1) < 2 {
		t.Fatal("missing Style 1 section in style-catalog.md")
	}
	classic := strings.Split(p1[1], "## 3. Style 2:")[0]

	anchors := extractPromptBlocks(classic)
	if len(anchors) != 2 {
		t.Fatalf("expected 2 anchors in Style 1 classic, got %d", len(anchors))
	}
	dark, light := anchors[0], anchors[1]

	expectedLight := strings.ReplaceAll(dark, "white line animation", "black line animation")
	expectedLight = strings.ReplaceAll(expectedLight, "figure stays white", "figure stays black")
	if light != expectedLight {
		t.Fatal("Theme swap must preserve the same anatomy and current-pose rule")
	}

	badAnatomyRe := regexp.MustCompile(`\d|one sixth|skeleton|rig definition|pen path`)
	for _, anchor := range anchors {
		for _, token := range []string{
			"hollow circular head",
			"stable proportions, connected line limbs, and uniform medium line weight",
			"Exactly two arms and two legs",
			"replaces that arm's previous line",
			"never leaving a resting arm beneath an active arm",
			"accent colors affect only props and graphic effects",
		} {
			if !strings.Contains(anchor, token) {
				t.Fatalf("missing required token in anchor: %s", token)
			}
		}
		if badAnatomyRe.MatchString(anchor) {
			t.Fatalf("unvalidated geometric recipe re-introduced in anchor: %s", anchor)
		}
	}

	for _, token := range []string{
		"Occasional simple line expressions are allowed by default.",
		"replace only \"no face, no eyes, no mouth,\"",
		"expression-enabled and left-handed variants were not separately tested",
		"not frame-by-frame certification",
	} {
		if !strings.Contains(catalog, token) {
			t.Fatalf("missing token in style-catalog.md: %s", token)
		}
	}

	for _, token := range []string{
		"repeat the concise frozen character anchor",
		"A raised arm replaces its old lowered line.",
	} {
		if !strings.Contains(contract, token) {
			t.Fatalf("missing token in omni-flash-prompt-contract.md: %s", token)
		}
	}
	if strings.Contains(contract, "no facial features, no hair, no clothing, no filled body, stable proportions, uniform medium line weight") {
		t.Fatal("forbidden obsolete prompt contract string present in contract")
	}

	if !strings.Contains(skillText, "For Style 1, freeze the concise character anchor") {
		t.Fatal("SKILL.md missing Style 1 concise character anchor instruction")
	}

	if _, err := os.Stat(filepath.Join(skill, "references", "continuity-lock.md")); err == nil {
		t.Fatal("references/continuity-lock.md exists; should be absent")
	}
}

type cozyFrame struct {
	Arms    int                       `json:"arms"`
	Legs    int                       `json:"legs"`
	Effects []string                  `json:"effects"`
	Props   map[string]cozyPropState  `json:"props"`
}

type cozyPropState struct {
	Locations   []string `json:"locations"`
	VisibleFace string   `json:"visible_face,omitempty"`
	ScreenDown  bool     `json:"screen_down,omitempty"`
}

type cozyClip struct {
	Scene           string                   `json:"scene"`
	CharacterAnchor string                   `json:"character_anchor"`
	VoiceProfile    string                   `json:"voice_profile"`
	PropDesigns     map[string]map[string]any`json:"prop_designs"`
	Frames          []cozyFrame              `json:"frames"`
	AllowedEffects  []string                 `json:"allowed_effects"`
}

type cozyPlan struct {
	SingleLocationRequested bool       `json:"single_location_requested"`
	ContinuingProps          []string   `json:"continuing_props"`
	Clips                    []cozyClip `json:"clips"`
}

func validateCozyPlan(plan cozyPlan) error {
	clips := plan.Clips
	if len(clips) == 0 {
		return fmt.Errorf("Missing clips")
	}
	if !plan.SingleLocationRequested {
		for i := 0; i < len(clips)-1; i++ {
			if clips[i].Scene == clips[i+1].Scene {
				return fmt.Errorf("New segments must establish different locations by default")
			}
		}
	}
	anchorSet := make(map[string]bool)
	voiceSet := make(map[string]bool)
	for _, c := range clips {
		anchorSet[c.CharacterAnchor] = true
		voiceSet[c.VoiceProfile] = true
	}
	if len(anchorSet) != 1 {
		return fmt.Errorf("Character design drift")
	}
	if len(voiceSet) != 1 {
		return fmt.Errorf("Voice identity drift")
	}

	for _, clip := range clips {
		for _, design := range clip.PropDesigns {
			if _, hasLoc := design["location"]; hasLoc {
				return fmt.Errorf("Initial prop positions do not belong in immutable designs")
			}
		}
		allowedEffects := make(map[string]bool)
		for _, eff := range clip.AllowedEffects {
			allowedEffects[eff] = true
		}
		departed := make(map[string]bool)
		for _, frame := range clip.Frames {
			if frame.Arms != 2 || frame.Legs != 2 {
				return fmt.Errorf("Extra anatomy")
			}
			for _, eff := range frame.Effects {
				if !allowedEffects[eff] {
					return fmt.Errorf("Effect leakage")
				}
			}
			for name, state := range frame.Props {
				if len(state.Locations) != 1 {
					return fmt.Errorf("Simultaneous prop copies")
				}
				location := state.Locations[0]
				if departed[name] {
					if location != "outside" {
						return fmt.Errorf("Departed prop restored in ending")
					}
				}
				if location == "outside" {
					departed[name] = true
				}
				if state.ScreenDown {
					if state.VisibleFace != "back" {
						return fmt.Errorf("Screen-down surface is wrong")
					}
				}
			}
		}
	}

	for i := 0; i < len(clips)-1; i++ {
		before := clips[i]
		after := clips[i+1]
		beforeLastFrame := before.Frames[len(before.Frames)-1]
		afterFirstFrame := after.Frames[0]
		for _, name := range plan.ContinuingProps {
			outgoing, ok1 := beforeLastFrame.Props[name]
			incoming, ok2 := afterFirstFrame.Props[name]
			if !ok1 || !ok2 || len(outgoing.Locations) == 0 || len(incoming.Locations) == 0 ||
				outgoing.Locations[0] != incoming.Locations[0] {
				return fmt.Errorf("Unplanned prop handoff")
			}
		}
	}
	return nil
}

func makeCozyFrame(phonePlace string, face string, screenDown bool, effects []string) cozyFrame {
	if face == "" {
		face = "front"
	}
	if effects == nil {
		effects = []string{}
	}
	return cozyFrame{
		Arms:    2,
		Legs:    2,
		Effects: effects,
		Props: map[string]cozyPropState{
			"phone": {
				Locations:   []string{phonePlace},
				VisibleFace: face,
				ScreenDown:  screenDown,
			},
		},
	}
}

func cloneCozyPlan(p cozyPlan) cozyPlan {
	data, _ := json.Marshal(p)
	var cp cozyPlan
	_ = json.Unmarshal(data, &cp)
	return cp
}

func TestVerifyCozyStorytelling(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")
	refPath := filepath.Join(skill, "references", "cozy-storytelling.md")
	refBytes, err := os.ReadFile(refPath)
	if err != nil {
		t.Fatalf("reading cozy-storytelling.md: %v", err)
	}
	ref := string(refBytes)

	voiceDocBytes, err := os.ReadFile(filepath.Join(skill, "references", "voice-continuity.md"))
	if err != nil {
		t.Fatalf("reading voice-continuity.md: %v", err)
	}
	voiceBlocks := extractPromptBlocks(string(voiceDocBytes))
	cozyBlocks := extractPromptBlocks(ref)
	if len(cozyBlocks) == 0 || len(voiceBlocks) == 0 {
		t.Fatal("missing blocks in cozy or voice docs")
	}
	anchor := cozyBlocks[0]
	voice := voiceBlocks[0]

	plan := cozyPlan{
		SingleLocationRequested: false,
		ContinuingProps:          []string{"phone"},
		Clips: []cozyClip{
			{
				Scene:           "office",
				CharacterAnchor: anchor,
				VoiceProfile:    voice,
				PropDesigns:     map[string]map[string]any{"phone": {"shape": "rectangle"}},
				Frames: []cozyFrame{
					makeCozyFrame("office_table", "front", false, nil),
					makeCozyFrame("right_hand", "front", false, nil),
				},
				AllowedEffects: []string{"red_circle"},
			},
			{
				Scene:           "platform",
				CharacterAnchor: anchor,
				VoiceProfile:    voice,
				PropDesigns:     map[string]map[string]any{"phone": {"shape": "rectangle"}},
				Frames: []cozyFrame{
					makeCozyFrame("right_hand", "front", false, nil),
					makeCozyFrame("right_hand", "front", false, nil),
				},
				AllowedEffects: []string{"red_circle"},
			},
			{
				Scene:           "outdoor_cafe",
				CharacterAnchor: anchor,
				VoiceProfile:    voice,
				PropDesigns:     map[string]map[string]any{"phone": {"shape": "rectangle"}},
				Frames: []cozyFrame{
					makeCozyFrame("right_hand", "front", false, nil),
					makeCozyFrame("cafe_table", "back", true, nil),
				},
				AllowedEffects: []string{"steam", "leaves"},
			},
		},
	}
	plan.Clips[0].Frames[0].Effects = []string{"red_circle"}
	plan.Clips[0].Frames[1].Effects = []string{"red_circle"}
	plan.Clips[1].Frames[0].Effects = []string{"red_circle"}
	plan.Clips[1].Frames[1].Effects = []string{"red_circle"}
	plan.Clips[2].Frames[0].Effects = []string{"steam", "leaves"}
	plan.Clips[2].Frames[1].Effects = []string{"steam", "leaves"}

	if err := validateCozyPlan(plan); err != nil {
		t.Fatalf("valid plan failed validation: %v", err)
	}

	single := cloneCozyPlan(plan)
	single.SingleLocationRequested = true
	for i := range single.Clips {
		single.Clips[i].Scene = "cafe"
	}
	if err := validateCozyPlan(single); err != nil {
		t.Fatalf("single location plan failed validation: %v", err)
	}

	// 8 bad cases
	// 1. repeated scene
	bad1 := cloneCozyPlan(plan)
	for i := range bad1.Clips {
		bad1.Clips[i].Scene = "same_desk"
	}
	if err := validateCozyPlan(bad1); err == nil {
		t.Fatal("expected repeated scene to fail")
	}

	// 2. duplicate prop location
	bad2 := cloneCozyPlan(plan)
	lastIdx := len(bad2.Clips[0].Frames) - 1
	pState := bad2.Clips[0].Frames[lastIdx].Props["phone"]
	pState.Locations = append(pState.Locations, "office_table")
	bad2.Clips[0].Frames[lastIdx].Props["phone"] = pState
	if err := validateCozyPlan(bad2); err == nil {
		t.Fatal("expected duplicate prop location to fail")
	}

	// 3. wrong face (screen down with front visible face)
	bad3 := cloneCozyPlan(plan)
	lastClipIdx := len(bad3.Clips) - 1
	lastFrameIdx := len(bad3.Clips[lastClipIdx].Frames) - 1
	pState3 := bad3.Clips[lastClipIdx].Frames[lastFrameIdx].Props["phone"]
	pState3.VisibleFace = "front"
	bad3.Clips[lastClipIdx].Frames[lastFrameIdx].Props["phone"] = pState3
	if err := validateCozyPlan(bad3); err == nil {
		t.Fatal("expected wrong face to fail")
	}

	// 4. departed prop restored
	bad4 := cloneCozyPlan(plan)
	bad4.Clips[lastClipIdx].Frames = append(bad4.Clips[lastClipIdx].Frames,
		makeCozyFrame("outside", "back", true, nil),
		makeCozyFrame("cafe_table", "back", true, nil),
	)
	if err := validateCozyPlan(bad4); err == nil {
		t.Fatal("expected departed prop restoration to fail")
	}

	// 5. extra arm
	bad5 := cloneCozyPlan(plan)
	bad5.Clips[1].Frames[0].Arms = 3
	if err := validateCozyPlan(bad5); err == nil {
		t.Fatal("expected extra arm to fail")
	}

	// 6. recast voice
	bad6 := cloneCozyPlan(plan)
	bad6.Clips[1].VoiceProfile = "Same narrator."
	if err := validateCozyPlan(bad6); err == nil {
		t.Fatal("expected recast voice to fail")
	}

	// 7. location in prop_designs
	bad7 := cloneCozyPlan(plan)
	bad7.Clips[0].PropDesigns["phone"]["location"] = "office_table"
	if err := validateCozyPlan(bad7); err == nil {
		t.Fatal("expected prop design location to fail")
	}

	// 8. effect leakage
	bad8 := cloneCozyPlan(plan)
	bad8.Clips[lastClipIdx].Frames[lastFrameIdx].Effects = []string{"red_circle"}
	if err := validateCozyPlan(bad8); err == nil {
		t.Fatal("expected effect leakage to fail")
	}

	// Doc checks
	for _, fn := range []string{
		"SKILL.md", "references/style-catalog.md",
		"references/storyboard-template.md", "references/omni-flash-prompt-contract.md",
	} {
		c, err := os.ReadFile(filepath.Join(skill, fn))
		if err != nil {
			t.Fatalf("reading %s: %v", fn, err)
		}
		if !strings.Contains(string(c), "cozy-storytelling.md") {
			t.Fatalf("missing cozy-storytelling.md in %s", fn)
		}
	}

	if !strings.Contains(ref, "voice-continuity.md") {
		t.Fatal("cozy-storytelling.md missing voice-continuity.md")
	}
	if !strings.Contains(ref, "Cozy Storytelling / 治愈故事风") {
		t.Fatal("cozy-storytelling.md missing style title")
	}

	// Media check
	for _, asset := range []string{"style-4-cozy-station", "style-4-cozy-cafe"} {
		gifPath := filepath.Join(root, "assets", "readme", asset+".gif")
		fi, err := os.Stat(gifPath)
		if err != nil || fi.IsDir() || fi.Size() >= 10*1024*1024 {
			t.Fatalf("missing or oversized gif: %s", gifPath)
		}
		gifBytes, err := os.ReadFile(gifPath)
		if err != nil {
			t.Fatalf("reading gif %s: %v", gifPath, err)
		}
		if len(gifBytes) < 6 || (!bytes.Equal(gifBytes[:6], []byte("GIF87a")) && !bytes.Equal(gifBytes[:6], []byte("GIF89a"))) {
			t.Fatalf("invalid gif header for %s", gifPath)
		}
		mp4Path := filepath.Join(root, "assets", "readme", asset+".mp4")
		if fi, err := os.Stat(mp4Path); err != nil || fi.IsDir() {
			t.Fatalf("missing mp4: %s", mp4Path)
		}
	}
}

func TestVerifyDocs(t *testing.T) {
	root := findRepoRoot(t)

	var paths []string
	paths = append(paths, filepath.Join(root, "README.md"))
	paths = append(paths, filepath.Join(root, "examples", "README.md"))

	docsMatches, err := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	if err != nil {
		t.Fatalf("glob docs/*.md: %v", err)
	}
	paths = append(paths, docsMatches...)

	catMatches, err := filepath.Glob(filepath.Join(root, "docs", "catalog", "*.md"))
	if err != nil {
		t.Fatalf("glob docs/catalog/*.md: %v", err)
	}
	paths = append(paths, catMatches...)

	mdLinkRe := regexp.MustCompile(`\]\(([^)]+)\)`)
	htmlLinkRe := regexp.MustCompile(`(?:href|src)="([^"]+)"`)
	anchorTagRe := regexp.MustCompile(`<a id="([^"]+)"></a>`)
	catalogDir := filepath.Join(root, "docs", "catalog")

	var errors []string
	for _, path := range paths {
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		content := string(contentBytes)

		var rawLinks []string
		for _, m := range mdLinkRe.FindAllStringSubmatch(content, -1) {
			rawLinks = append(rawLinks, m[1])
		}
		for _, m := range htmlLinkRe.FindAllStringSubmatch(content, -1) {
			rawLinks = append(rawLinks, m[1])
		}

		for _, rawLink := range rawLinks {
			u, err := url.Parse(rawLink)
			if err != nil || u.Scheme != "" || u.Host != "" {
				continue
			}
			var target string
			if u.Path != "" {
				decodedPath, err := url.PathUnescape(u.Path)
				if err != nil {
					decodedPath = u.Path
				}
				target = filepath.Clean(filepath.Join(filepath.Dir(path), decodedPath))
			} else {
				target = path
			}

			if fi, err := os.Stat(target); err != nil || fi.IsDir() {
				rel, _ := filepath.Rel(root, path)
				errors = append(errors, fmt.Sprintf("%s: missing %s", rel, rawLink))
				continue
			}

			if u.Fragment != "" && filepath.Dir(target) == catalogDir {
				targetBytes, err := os.ReadFile(target)
				if err != nil {
					t.Fatalf("reading %s: %v", target, err)
				}
				targetAnchors := anchorTagRe.FindAllStringSubmatch(string(targetBytes), -1)
				found := false
				decodedFrag, _ := url.PathUnescape(u.Fragment)
				for _, a := range targetAnchors {
					if a[1] == decodedFrag || a[1] == u.Fragment {
						found = true
						break
					}
				}
				if !found {
					rel, _ := filepath.Rel(root, path)
					errors = append(errors, fmt.Sprintf("%s: missing anchor %s", rel, rawLink))
				}
			}
		}
	}

	if len(errors) > 0 {
		t.Fatalf("doc verification errors:\n%s", strings.Join(errors, "\n"))
	}

	catalogReadmes, err := filepath.Glob(filepath.Join(catalogDir, "README.*.md"))
	if err != nil {
		t.Fatalf("glob catalog readmes: %v", err)
	}
	readmeNames := make(map[string]bool)
	for _, cr := range catalogReadmes {
		readmeNames[filepath.Base(cr)] = true
	}
	if len(readmeNames) != 2 || !readmeNames["README.en.md"] || !readmeNames["README.tr.md"] {
		t.Fatalf("expected {README.en.md, README.tr.md}, got %v", readmeNames)
	}
}

func getJPEGDimensions(data []byte) (int, int, error) {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 0, 0, fmt.Errorf("not a jpeg")
	}
	offset := 2
	for offset < len(data) {
		if data[offset] != 0xFF {
			return 0, 0, fmt.Errorf("expected 0xFF at offset %d", offset)
		}
		for offset < len(data) && data[offset] == 0xFF {
			offset++
		}
		if offset >= len(data) {
			break
		}
		marker := data[offset]
		offset++
		if marker == 0xD9 || marker == 0xDA {
			break
		}
		if offset+2 > len(data) {
			return 0, 0, fmt.Errorf("unexpected EOF reading length")
		}
		length := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		if marker == 0xC0 || marker == 0xC1 || marker == 0xC2 {
			if offset+7 > len(data) {
				return 0, 0, fmt.Errorf("unexpected EOF reading SOF")
			}
			height := int(binary.BigEndian.Uint16(data[offset+3 : offset+5]))
			width := int(binary.BigEndian.Uint16(data[offset+5 : offset+7]))
			return width, height, nil
		}
		offset += length
	}
	return 0, 0, fmt.Errorf("SOF marker not found")
}

func TestVerifyThumbnailKit(t *testing.T) {
	root := findRepoRoot(t)

	enSourceBytes, err := os.ReadFile(filepath.Join(root, "docs", "catalog", "README.en.md"))
	if err != nil {
		t.Fatalf("reading README.en.md: %v", err)
	}
	source := string(enSourceBytes)

	start := "<!-- thumbnail-publishing-kit:start -->"
	end := "<!-- thumbnail-publishing-kit:end -->"
	if strings.Count(source, start) != 1 || strings.Count(source, end) != 1 {
		t.Fatal("start or end marker count is not 1 in README.en.md")
	}

	kit := strings.Split(strings.Split(source, start)[1], end)[0]
	if !strings.Contains(kit, "## 🖼️ High-CTR Thumbnail & Publishing Kit") {
		t.Fatal("missing heading in kit")
	}

	idxB2 := strings.Index(source, "<a id=\"b2\"></a>")
	idxStart := strings.Index(source, start)
	idxFaq := strings.Index(source, "<a id=\"faq\"></a>")
	if !(idxB2 < idxStart && idxStart < idxFaq) {
		t.Fatal("kit placement is not between b2 and faq")
	}

	if strings.Contains(source, "[Publishing Kit](#thumbnail-publishing-kit)") {
		t.Fatal("found unwanted [Publishing Kit](#thumbnail-publishing-kit) in source")
	}
	if !strings.Contains(source, "[FAQ](#faq) · [Roadmap](#road-to-10000-stars--creator-ecosystem)") {
		t.Fatal("missing FAQ/Roadmap link in source")
	}

	for _, token := range []string{
		"**Your video is only as good as the click.**",
		"both **Director and Autopilot Mode**",
		"*All thumbnail concepts are designed around proven YouTube packaging psychology:",
		"### What the kit delivers:",
		"### How to trigger it:",
		"After you confirm the final video",
		"No need to copy a prompt.",
		"An early request based on an approved script produces drafts",
		"nothing is uploaded or published automatically.",
		"Content-ID clearance and monetization eligibility are not guaranteed.",
	} {
		if !strings.Contains(kit, token) {
			t.Fatalf("missing required token in kit: %s", token)
		}
	}

	deliverablesRe := regexp.MustCompile(`(?m)^- \*\*([^*]+)\*\*:`)
	var deliverables []string
	for _, m := range deliverablesRe.FindAllStringSubmatch(kit, -1) {
		deliverables = append(deliverables, m[1])
	}
	expectedDeliverables := []string{
		"Thumbnail Image", "5 High-CTR Title Variants",
		"YouTube Studio Description & SEO", "Pinned Comment Hook", "Pre-Publish Checklist",
	}
	if len(deliverables) != len(expectedDeliverables) {
		t.Fatalf("unexpected deliverables count: got %v", deliverables)
	}
	for i, d := range deliverables {
		if d != expectedDeliverables[i] {
			t.Fatalf("deliverable %d mismatch: got %s, want %s", i, d, expectedDeliverables[i])
		}
	}

	promptBlock := "```text\n" +
		"Generate the Thumbnail & Publishing Kit for this video:\n" +
		"- 1 landscape 16:9 thumbnail image matching the approved art style\n" +
		"- 5 YouTube title options\n" +
		"- A description, relevant tags, and verified timestamps if available\n" +
		"- 1 engagement question for the pinned comment\n" +
		"- A pre-publish checklist with any pending checks clearly marked\n" +
		"```"
	if !strings.Contains(kit, promptBlock) {
		t.Fatal("missing prompt block in kit")
	}

	if strings.Contains(source, "Copy the image prompts") || strings.Contains(source, "docs/thumbnail-image-prompts.md") {
		t.Fatal("obsolete prompt text found in source")
	}

	imageSlugs := []string{
		"no-camera", "first-job", "a5-samurai-taxes",
		"a1-you-vs-you", "a2-busy-vs-better", "a4-3am-again",
	}
	var imagePaths []string
	var imageURLs []string
	var expectedHrefs []string
	for _, slug := range imageSlugs {
		p := fmt.Sprintf("assets/readme/thumbnails/%s.jpg", slug)
		imagePaths = append(imagePaths, p)
		imageURLs = append(imageURLs, "https://cdn.jsdelivr.net/gh/egnake/Tubie-Stickman@ae1685ba812f9686463f319fe6f0d4eb7d581433/"+p)
		expectedHrefs = append(expectedHrefs, "../../"+p)
	}

	imgSrcRe := regexp.MustCompile(`<img src="([^"]+)"`)
	var foundImgs []string
	for _, m := range imgSrcRe.FindAllStringSubmatch(kit, -1) {
		foundImgs = append(foundImgs, m[1])
	}
	if len(foundImgs) != len(imageURLs) {
		t.Fatalf("img src count mismatch: got %d, want %d", len(foundImgs), len(imageURLs))
	}
	for i := range foundImgs {
		if foundImgs[i] != imageURLs[i] {
			t.Fatalf("img src %d mismatch: got %s, want %s", i, foundImgs[i], imageURLs[i])
		}
	}

	aHrefRe := regexp.MustCompile(`<a href="([^"]+\.jpg)"`)
	var foundHrefs []string
	for _, m := range aHrefRe.FindAllStringSubmatch(kit, -1) {
		foundHrefs = append(foundHrefs, m[1])
	}
	if len(foundHrefs) != len(expectedHrefs) {
		t.Fatalf("a href count mismatch: got %d, want %d", len(foundHrefs), len(expectedHrefs))
	}
	for i := range foundHrefs {
		if foundHrefs[i] != expectedHrefs[i] {
			t.Fatalf("a href %d mismatch: got %s, want %s", i, foundHrefs[i], expectedHrefs[i])
		}
	}

	altRe := regexp.MustCompile(`<img[^>]+alt="[^"]+"`)
	if len(altRe.FindAllString(kit, -1)) != 6 {
		t.Fatal("alt tag count in kit is not 6")
	}
	if strings.Count(kit, "<tr>") != 4 || strings.Count(kit, "</tr>") != 4 {
		t.Fatal("<tr> count in kit is not 4")
	}
	if strings.Count(kit, "<table width=\"100%\">") != 2 || strings.Count(kit, "</table>") != 2 {
		t.Fatal("table count in kit is not 2")
	}
	if regexp.MustCompile(`<th\b`).MatchString(kit) {
		t.Fatal("found <th in kit")
	}

	for _, p := range imagePaths {
		data, err := os.ReadFile(filepath.Join(root, p))
		if err != nil {
			t.Fatalf("reading %s: %v", p, err)
		}
		if !bytes.HasPrefix(data, []byte("\xff\xd8")) || !bytes.HasSuffix(data, []byte("\xff\xd9")) {
			t.Fatalf("SOI/EOI marker missing in %s", p)
		}
		if len(data) >= 2000000 {
			t.Fatalf("%s exceeds 2MB: %d", p, len(data))
		}
		w, h, err := getJPEGDimensions(data)
		if err != nil {
			t.Fatalf("reading jpeg dimensions for %s: %v", p, err)
		}
		if w != 1280 || h != 720 {
			t.Fatalf("dimensions for %s are %dx%d, want 1280x720", p, w, h)
		}
	}

	roadmapLabels := map[string]string{"en": "Roadmap", "tr": "Yol haritası"}
	for locale, roadmap := range roadmapLabels {
		fn := fmt.Sprintf("README.%s.md", locale)
		catBytes, err := os.ReadFile(filepath.Join(root, "docs", "catalog", fn))
		if err != nil {
			t.Fatalf("reading %s: %v", fn, err)
		}
		text := string(catBytes)
		if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 {
			t.Fatalf("marker count in %s is not 1", fn)
		}
		finishingLink := strings.Split(text, "<a id=\"style-gallery\"></a>")[0]
		if !regexp.MustCompile(`\]\(#thumbnail-publishing-kit\)`).MatchString(finishingLink) {
			t.Fatalf("missing finishing link in %s", fn)
		}
		localized := strings.Split(strings.Split(text, start)[1], end)[0]
		idxB2Loc := strings.Index(text, "<a id=\"b2\"></a>")
		idxStartLoc := strings.Index(text, start)
		idxFaqLoc := strings.Index(text, "<a id=\"faq\"></a>")
		if !(idxB2Loc < idxStartLoc && idxStartLoc < idxFaqLoc) {
			t.Fatalf("localized kit placement error in %s", fn)
		}

		var locImgs []string
		for _, m := range imgSrcRe.FindAllStringSubmatch(localized, -1) {
			locImgs = append(locImgs, m[1])
		}
		if len(locImgs) != len(imageURLs) {
			t.Fatalf("localized img count mismatch in %s", fn)
		}

		var locHrefs []string
		for _, m := range aHrefRe.FindAllStringSubmatch(localized, -1) {
			locHrefs = append(locHrefs, m[1])
		}
		if len(locHrefs) != len(expectedHrefs) {
			t.Fatalf("localized href count mismatch in %s", fn)
		}

		if len(deliverablesRe.FindAllString(localized, -1)) != 5 {
			t.Fatalf("deliverables count in %s is not 5", fn)
		}
		if len(regexp.MustCompile(`(?m)^### `).FindAllString(localized, -1)) != 2 {
			t.Fatalf("h3 count in %s is not 2", fn)
		}
		if strings.Count(localized, "<table width=\"100%\">") != 2 || strings.Count(localized, "</table>") != 2 {
			t.Fatalf("table count in %s is not 2", fn)
		}
		if strings.Count(localized, "<tr>") != 4 || strings.Count(localized, "</tr>") != 4 {
			t.Fatalf("tr count in %s is not 4", fn)
		}
		if len(regexp.MustCompile(`<td align="center">([^<]+)</td>`).FindAllStringSubmatch(localized, -1)) != 6 {
			t.Fatalf("td count in %s is not 6", fn)
		}
		if regexp.MustCompile(`<th\b`).MatchString(localized) {
			t.Fatalf("found th in %s", fn)
		}

		locPrompts := extractPromptBlocks(localized)
		if len(locPrompts) != 1 || len(regexp.MustCompile(`(?m)^- `).FindAllString(locPrompts[0], -1)) != 5 {
			t.Fatalf("prompt block format error in %s", fn)
		}
		for _, tok := range []string{"16:9", "1", "5", "YouTube"} {
			if !strings.Contains(locPrompts[0], tok) {
				t.Fatalf("missing prompt token %s in %s", tok, fn)
			}
		}
		for _, tok := range []string{"YouTube Studio", "SEO", "Content-ID", "TikTok"} {
			if !strings.Contains(localized, tok) {
				t.Fatalf("missing localized token %s in %s", tok, fn)
			}
		}

		var navLine string
		for _, line := range strings.Split(text, "\n") {
			if strings.Contains(line, "](#quick-start)") && strings.Contains(line, "](#faq)") {
				navLine = line
				break
			}
		}
		if navLine == "" {
			t.Fatalf("missing nav line in %s", fn)
		}
		navTargetRe := regexp.MustCompile(`\]\((#[^)]+)\)`)
		var navTargets []string
		for _, m := range navTargetRe.FindAllStringSubmatch(navLine, -1) {
			navTargets = append(navTargets, m[1])
		}
		expectedNavTargets := []string{
			"#quick-start", "#style-gallery", "#faq", "#road-to-10000-stars--creator-ecosystem", "#community",
		}
		if len(navTargets) != len(expectedNavTargets) {
			t.Fatalf("nav target count mismatch in %s: %v", fn, navTargets)
		}
		for i := range navTargets {
			if navTargets[i] != expectedNavTargets[i] {
				t.Fatalf("nav target %d mismatch in %s: got %s, want %s", i, fn, navTargets[i], expectedNavTargets[i])
			}
		}
		if !strings.Contains(navLine, fmt.Sprintf("[%s](#road-to-10000-stars--creator-ecosystem)", roadmap)) ||
			!strings.HasSuffix(navLine, "](#community)") {
			t.Fatalf("nav line suffix or roadmap label error in %s: %s", fn, navLine)
		}
		if locale != "en" {
			if strings.Contains(localized, "Your video is only as good as the click.") ||
				strings.Contains(localized, "### What the kit delivers:") ||
				strings.Contains(localized, "A simple question and one clear symbol.") {
				t.Fatalf("English text leaked in %s", fn)
			}
		}
	}

	skill := filepath.Join(root, "skills", "directing-stickman-videos")
	entryBytes, _ := os.ReadFile(filepath.Join(skill, "SKILL.md"))
	guideBytes, _ := os.ReadFile(filepath.Join(skill, "references", "getting-started.md"))
	autopilotBytes, _ := os.ReadFile(filepath.Join(skill, "references", "autopilot.md"))
	finishingPath := filepath.Join(skill, "references", "thumbnail-publishing-kit.md")
	finishingBytes, err := os.ReadFile(finishingPath)
	if err != nil {
		t.Fatalf("reading %s: %v", finishingPath, err)
	}

	if !strings.Contains(string(entryBytes), "(references/thumbnail-publishing-kit.md)") {
		t.Fatal("SKILL.md missing thumbnail kit link")
	}
	if !strings.Contains(string(guideBytes), "(thumbnail-publishing-kit.md)") ||
		!strings.Contains(string(autopilotBytes), "(thumbnail-publishing-kit.md)") {
		t.Fatal("guide or autopilot missing thumbnail kit link")
	}

	mdLinkRe := regexp.MustCompile(`\]\(([^)]+\.md)(?:#[^)]*)?\)`)
	for _, doc := range []struct {
		name string
		text string
		dir  string
	}{
		{"SKILL.md", string(entryBytes), skill},
		{"getting-started.md", string(guideBytes), filepath.Join(skill, "references")},
		{"autopilot.md", string(autopilotBytes), filepath.Join(skill, "references")},
		{"thumbnail-publishing-kit.md", string(finishingBytes), filepath.Join(skill, "references")},
	} {
		for _, m := range mdLinkRe.FindAllStringSubmatch(doc.text, -1) {
			target := m[1]
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
				continue
			}
			full := filepath.Join(doc.dir, target)
			if fi, err := os.Stat(full); err != nil || fi.IsDir() {
				t.Fatalf("broken link in %s: %s -> %s", doc.name, target, full)
			}
		}
	}

	numberedRe := regexp.MustCompile(`(?m)^\d\.\s\*\*`)
	if len(numberedRe.FindAllString(string(finishingBytes), -1)) != 5 {
		t.Fatal("thumbnail-publishing-kit.md does not contain 5 numbered items")
	}
}

func TestVerifyReadmes(t *testing.T) {
	root := findRepoRoot(t)

	for _, stem := range []string{
		"light-theme-demo", "dark-theme-demo", "style-2a-demo", "style-2b-demo",
		"style-3-demo", "style-3-community-fire", "style-4-cozy-station", "style-4-cozy-cafe",
	} {
		for _, ext := range []string{"gif", "mp4"} {
			p := filepath.Join(root, "assets", "readme", fmt.Sprintf("%s.%s", stem, ext))
			fi, err := os.Stat(p)
			if err != nil || fi.IsDir() || fi.Size() == 0 {
				t.Fatalf("missing or empty demo asset: %s", p)
			}
		}
	}

	for _, stem := range []string{"style-5-japan", "style-5-egypt"} {
		p := filepath.Join(root, "assets", "readme", fmt.Sprintf("%s.mp4", stem))
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() || fi.Size() == 0 {
			t.Fatalf("missing or empty historical demo asset: %s", p)
		}
	}
}

func TestVerifyA6Guidance(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")

	entryBytes, err := os.ReadFile(filepath.Join(skill, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}
	entry := string(entryBytes)

	guideBytes, err := os.ReadFile(filepath.Join(skill, "references", "getting-started.md"))
	if err != nil {
		t.Fatalf("reading getting-started.md: %v", err)
	}
	guide := string(guideBytes)

	refPath := filepath.Join(skill, "references", "amber-noir.md")
	refBytes, err := os.ReadFile(refPath)
	if err != nil {
		t.Fatalf("reading amber-noir.md: %v", err)
	}
	a6 := string(refBytes)

	if !strings.Contains(entry, "(references/amber-noir.md)") ||
		!strings.Contains(entry, "when selected, before planning") ||
		!strings.Contains(entry, "bypass the A1–A5 contract below") ||
		!strings.Contains(entry, "apply only to A1–A5, not to A6 or any B style") {
		t.Fatal("SKILL.md missing A6 contract entries")
	}

	if !strings.Contains(guide, "(amber-noir.md)") {
		t.Fatal("getting-started.md missing amber-noir.md link")
	}

	wParts := strings.Split(guide, "### 欢迎使用火柴人视频 Skill\n")
	if len(wParts) < 2 {
		t.Fatal("missing Chinese welcome in guide")
	}
	welcome := strings.Split(wParts[1], "### Adapt instead of replaying the README")[0]
	if strings.Contains(welcome, "Renoise") || strings.Contains(welcome, "Veo 3.1 Fast") {
		t.Fatal("unwanted internal strings in welcome")
	}

	for _, token := range []string{
		"Explain only after selection",
		"Google Flow · Veo 3.1 Fast · 720p · up to 8 seconds per clip",
		"historical test provenance",
		"Route 1 · Flow browser (recommended)",
		"Route 2 · Renoise MCP (explicit choice)",
		"not permission to submit a generation",
		"not a proven mouth-stability fix",
		"independent Gemini Puck narration",
		"a6-finishing.md",
		"large uppercase outlined captions",
		"No creator watermark by default",
		"https://renoise.ai/mcp?utm_source=crazyegnake-aini-202610&utm_medium=xcom",
		"login-only Flow project URL is not a usable public media reference",
		"MCP installation alone is not proof",
		"do not split mechanically by word count",
		"at least four clips", "at least eight",
		"Do not repaint, mask, replace the mouth",
		"Do not promise that every Codex/Claude environment has browser control",
		"No upload/publication follows automatically",
	} {
		if !strings.Contains(a6, token) {
			t.Fatalf("missing required token in amber-noir.md: %s", token)
		}
	}

	if regexp.MustCompile(`/Users/|/home/|70847984-|task:366834`).MatchString(a6) {
		t.Fatal("forbidden user path or task ID in amber-noir.md")
	}

	contractPath := filepath.Join(skill, "references", "a6-reference-prompts.md")
	notesPath := filepath.Join(skill, "assets", "a6", "SOURCE-NOTES.md")
	linkRe := regexp.MustCompile(`\]\(([^)]+)\)`)
	for _, docPath := range []string{refPath, contractPath, notesPath} {
		b, err := os.ReadFile(docPath)
		if err != nil {
			t.Fatalf("reading %s: %v", docPath, err)
		}
		text := string(b)
		if regexp.MustCompile(`/Users/|/home/`).MatchString(text) {
			t.Fatalf("user path in %s", docPath)
		}
		for _, m := range linkRe.FindAllStringSubmatch(text, -1) {
			target := m[1]
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "#") {
				continue
			}
			targetClean := strings.Split(target, "#")[0]
			full := filepath.Clean(filepath.Join(filepath.Dir(docPath), targetClean))
			if fi, err := os.Stat(full); err != nil || fi.IsDir() {
				t.Fatalf("broken link in %s: %s -> %s", docPath, target, full)
			}
		}
	}

	for _, required := range []string{
		"reference-manifest.json", "a6-reference-prompts.md", "a6_reference_check.go",
		"Before claiming that references are missing", "not a visual-quality verdict",
	} {
		if !strings.Contains(a6, required) {
			t.Fatalf("missing required string in amber-noir.md: %s", required)
		}
	}

	// Verify root READMEs do not leak internal strings
	rootReadmes, err := filepath.Glob(filepath.Join(root, "README*.md"))
	if err != nil {
		t.Fatalf("glob root readmes: %v", err)
	}
	for _, rPath := range rootReadmes {
		b, err := os.ReadFile(rPath)
		if err != nil {
			t.Fatalf("reading %s: %v", rPath, err)
		}
		text := string(b)
		if strings.Contains(text, "Route 2 · Renoise MCP (explicit choice)") ||
			strings.Contains(text, "utm_source=crazyegnake-aini-202610") {
			t.Fatalf("internal A6 strings leaked in %s", rPath)
		}
	}

	// a6-finishing.md
	finishingBytes, err := os.ReadFile(filepath.Join(skill, "references", "a6-finishing.md"))
	if err != nil {
		t.Fatalf("reading a6-finishing.md: %v", err)
	}
	finishing := string(finishingBytes)
	for _, tok := range []string{
		"Gemini TTS · Puck", "curious, exploratory", "ALL CAPS",
		"pure white", "No `@CrazyKaomei`", "a6_finish.go",
		"not a command to reprocess approved audio",
	} {
		if !strings.Contains(finishing, tok) {
			t.Fatalf("missing token in a6-finishing.md: %s", tok)
		}
	}
	if fi, err := os.Stat(filepath.Join(skill, "scripts", "a6_finish.go")); err != nil || fi.IsDir() {
		t.Fatal("missing skills/.../scripts/a6_finish.go")
	}

	// Run a6_reference_check.go on pack and fixtures
	runRefCheck := func(args ...string) {
		cmdArgs := append([]string{"run", "scripts/a6_reference_check.go"}, args...)
		cmd := exec.Command("go", cmdArgs...)
		cmd.Dir = skill
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("a6_reference_check.go failed (%v):\n%s", err, string(out))
		}
	}
	runRefCheck()
	runRefCheck("-plan", "tests/fixtures/a6-night-library.plan.json")
	runRefCheck("-plan", "tests/fixtures/a6-coastal-station.plan.json")
}

func TestVerifyA6Finishing(t *testing.T) {
	root := findRepoRoot(t)
	skill := filepath.Join(root, "skills", "directing-stickman-videos")

	cmd := exec.Command("go", "test", "-v", "./tests", "-run", "TestA6Finish")
	cmd.Dir = skill
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test TestA6Finish failed (%v):\n%s", err, string(out))
	}
}
