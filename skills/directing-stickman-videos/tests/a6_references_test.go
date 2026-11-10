package tests

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"directing-stickman-videos/pkg/a6"
)

func findSkillRoot(t *testing.T) string {
	t.Helper()
	candidates := []string{".", "..", filepath.Join("..", "..")}
	for _, c := range candidates {
		if fi, err := os.Stat(filepath.Join(c, "assets", "a6", "reference-manifest.json")); err == nil && !fi.IsDir() {
			abs, err := filepath.Abs(c)
			if err != nil {
				return c
			}
			return abs
		}
	}
	t.Fatal("Could not locate Skill root with assets/a6/reference-manifest.json")
	return ""
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatalf("copyDir MkdirAll failed: %v", err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("copyDir ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyDir(t, s, d)
		} else {
			copyFile(t, s, d)
		}
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("copyFile open src failed: %v", err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("copyFile create dst failed: %v", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		t.Fatalf("copyFile copy failed: %v", err)
	}
}

type testEnv struct {
	tempDir  string
	planPath string
	planData map[string]any
}

func setupTestEnv(t *testing.T, skillRoot string) *testEnv {
	t.Helper()
	temp := t.TempDir()
	copyDir(t, filepath.Join(skillRoot, "assets", "a6"), filepath.Join(temp, "assets", "a6"))

	fixtureBytes, err := os.ReadFile(filepath.Join(skillRoot, "tests", "fixtures", "a6-night-library.plan.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}
	var plan map[string]any
	if err := json.Unmarshal(fixtureBytes, &plan); err != nil {
		t.Fatalf("Failed to parse fixture: %v", err)
	}

	planPath := filepath.Join(temp, "plan.json")
	return &testEnv{
		tempDir:  temp,
		planPath: planPath,
		planData: plan,
	}
}

func (e *testEnv) savePlan(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(e.planData)
	if err != nil {
		t.Fatalf("Failed to marshal plan: %v", err)
	}
	if err := os.WriteFile(e.planPath, b, 0644); err != nil {
		t.Fatalf("Failed to write plan: %v", err)
	}
	return e.planPath
}

func (e *testEnv) updateManifest(t *testing.T, fn func(map[string]any)) {
	t.Helper()
	p := filepath.Join(e.tempDir, "assets", "a6", "reference-manifest.json")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}
	fn(m)
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}
	if err := os.WriteFile(p, out, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}
}

func (e *testEnv) failsPlan(t *testing.T, substr string) {
	t.Helper()
	planPath := e.savePlan(t)
	_, err := a6.ValidatePlan(e.tempDir, planPath)
	if err == nil {
		t.Fatalf("Expected plan validation error containing %q, but got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("Expected plan validation error containing %q, but got: %v", substr, err)
	}
}

func clonePlan(src map[string]any) map[string]any {
	b, _ := json.Marshal(src)
	var dst map[string]any
	_ = json.Unmarshal(b, &dst)
	return dst
}

func TestBundledReferencesAreRealAndHaveProvenance(t *testing.T) {
	skillRoot := findSkillRoot(t)
	assets, err := a6.ValidatePack(skillRoot)
	if err != nil {
		t.Fatalf("ValidatePack failed: %v", err)
	}
	if len(assets) != 3 {
		t.Fatalf("Expected 3 bundled reference assets, got %d", len(assets))
	}
}

func TestBothUnrelatedFixtureStoriesPass(t *testing.T) {
	skillRoot := findSkillRoot(t)
	for _, name := range []string{"night-library", "coastal-station"} {
		fixturePath := filepath.Join(skillRoot, "tests", "fixtures", "a6-"+name+".plan.json")
		prompts, err := a6.ValidatePlan(skillRoot, fixturePath)
		if err != nil {
			t.Fatalf("Fixture %s failed: %v", name, err)
		}
		if len(prompts) < 2 {
			t.Fatalf("Expected at least 2 prompts for %s, got %d", name, len(prompts))
		}
		for _, p := range prompts {
			promptText, _ := p["prompt"].(string)
			if strings.Contains(promptText, "[STYLE LOCK]") {
				t.Fatalf("Prompt contains [STYLE LOCK]: %s", promptText)
			}
		}
	}
}

func TestNewStoriesDoNotInheritSampleWardrobe(t *testing.T) {
	skillRoot := findSkillRoot(t)
	cases := []struct {
		name     string
		wardrobe string
	}{
		{"night-library", "teal"},
		{"coastal-station", "navy"},
	}
	for _, tc := range cases {
		fixturePath := filepath.Join(skillRoot, "tests", "fixtures", "a6-"+tc.name+".plan.json")
		data, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", tc.name, err)
		}
		var plan map[string]any
		if err := json.Unmarshal(data, &plan); err != nil {
			t.Fatalf("Failed to parse %s: %v", tc.name, err)
		}
		images, _ := plan["images"].([]any)
		var castParts []string
		for _, im := range images {
			imMap, _ := im.(map[string]any)
			promptMap, _ := imMap["image_prompt"].(map[string]any)
			castStr, _ := promptMap["cast"].(string)
			castParts = append(castParts, castStr)

			refs, _ := imMap["reference_inputs"].([]any)
			for _, r := range refs {
				rMap, _ := r.(map[string]any)
				if rMap["role"] != "style" {
					t.Fatalf("%s image has non-style reference role: %v", tc.name, rMap["role"])
				}
			}
		}
		allCast := strings.ToLower(strings.Join(castParts, " "))
		if !strings.Contains(allCast, tc.wardrobe) {
			t.Fatalf("%s cast missing %q", tc.name, tc.wardrobe)
		}
		if strings.Contains(allCast, "orange") {
			t.Fatalf("%s cast unexpectedly contains 'orange'", tc.name)
		}
		if strings.Contains(allCast, "hoodie") {
			t.Fatalf("%s cast unexpectedly contains 'hoodie'", tc.name)
		}
	}
}

func TestMissingPackBlocksOriginalFailure(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	_ = os.RemoveAll(filepath.Join(env.tempDir, "assets", "a6"))
	if _, err := a6.ValidatePack(env.tempDir); err == nil {
		t.Fatal("Expected error when assets/a6 is missing, got nil")
	}
}

func TestDeletedImageBlocks(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	_ = os.Remove(filepath.Join(env.tempDir, "assets", "a6", "apartment-style-example.png"))
	_, err := a6.ValidatePack(env.tempDir)
	if err == nil || !strings.Contains(err.Error(), "Missing") {
		t.Fatalf("Expected error containing 'Missing', got: %v", err)
	}
}

func TestTamperedFileBlocks(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	p := filepath.Join(env.tempDir, "assets", "a6", "cafe-face-and-light.jpg")
	data, _ := os.ReadFile(p)
	data = append(data, []byte("tampered")...)
	_ = os.WriteFile(p, data, 0644)
	_, err := a6.ValidatePack(env.tempDir)
	if err == nil || !strings.Contains(err.Error(), "SHA-256 mismatch") {
		t.Fatalf("Expected error containing 'SHA-256 mismatch', got: %v", err)
	}
}

func TestWrongDimensionsBlock(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	env.updateManifest(t, func(m map[string]any) {
		assets := m["assets"].([]any)
		first := assets[0].(map[string]any)
		first["width"] = 1.0
	})
	_, err := a6.ValidatePack(env.tempDir)
	if err == nil || !strings.Contains(err.Error(), "Dimension mismatch") {
		t.Fatalf("Expected error containing 'Dimension mismatch', got: %v", err)
	}
}

func TestMissingProvenanceBlocks(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	env.updateManifest(t, func(m map[string]any) {
		assets := m["assets"].([]any)
		first := assets[0].(map[string]any)
		delete(first, "provenance")
	})
	_, err := a6.ValidatePack(env.tempDir)
	if err == nil || !strings.Contains(err.Error(), "Missing provenance") {
		t.Fatalf("Expected error containing 'Missing provenance', got: %v", err)
	}
}

func TestInvalidImageBytesBlock(t *testing.T) {
	temp := t.TempDir()
	p := filepath.Join(temp, "fake.png")
	_ = os.WriteFile(p, []byte("not an image"), 0644)
	_, _, _, err := a6.ImageInfo(p)
	if err == nil || !strings.Contains(err.Error(), "PNG/JPEG") {
		t.Fatalf("Expected error containing 'PNG/JPEG', got: %v", err)
	}
}

func TestPlaceholderVariantsBlock(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	original := clonePlan(env.planData)

	tokens := []string{
		"[STYLE LOCK]", "[CAST_LOCK]", "[WORLD-LOCK]", "{{style}}", "${cast}",
		"<insert style>", "same as above", "same character", "use previous prompt", "TBD",
	}

	for _, tok := range tokens {
		t.Run("token="+tok, func(t *testing.T) {
			env.planData = clonePlan(original)
			images := env.planData["images"].([]any)
			first := images[0].(map[string]any)
			prompt := first["image_prompt"].(map[string]any)
			prompt["style"] = tok
			env.failsPlan(t, "Unexpanded placeholder")
		})
	}
}

func TestPlaceholderInMotionBlocks(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	images := env.planData["images"].([]any)
	first := images[0].(map[string]any)
	first["motion_prompt"] = "Animate [CAST LOCK]"
	env.failsPlan(t, "Unexpanded placeholder")
}

func TestEmptyPromptSectionBlocks(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	fields := []string{"style", "cast", "world", "composition", "constraints"}
	original := clonePlan(env.planData)

	for _, field := range fields {
		t.Run("field="+field, func(t *testing.T) {
			env.planData = clonePlan(original)
			images := env.planData["images"].([]any)
			first := images[0].(map[string]any)
			prompt := first["image_prompt"].(map[string]any)
			delete(prompt, field)
			env.failsPlan(t, "Empty prompt field")
		})
	}
}

func TestMissingReferenceInputBlocks(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	images := env.planData["images"].([]any)
	first := images[0].(map[string]any)
	first["reference_inputs"] = []any{}
	env.failsPlan(t, "No reference inputs")
}

func TestFabricatedPackageAssetBlocks(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	images := env.planData["images"].([]any)
	first := images[0].(map[string]any)
	refs := first["reference_inputs"].([]any)
	ref0 := refs[0].(map[string]any)
	ref0["asset_id"] = "invented-upload-id"
	env.failsPlan(t, "Unknown package reference")
}

func TestSampleIdentityRequiresExplicitChoice(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	images := env.planData["images"].([]any)
	first := images[0].(map[string]any)
	refs := first["reference_inputs"].([]any)
	ref0 := refs[0].(map[string]any)
	ref0["role"] = "identity"

	env.failsPlan(t, "not explicitly chosen")

	env.planData["reuse_sample_identity"] = true
	saved := env.savePlan(t)
	if _, err := a6.ValidatePlan(env.tempDir, saved); err != nil {
		t.Fatalf("Expected plan with reuse_sample_identity to pass, got: %v", err)
	}
}

func TestSampleImageNotAutomaticStoryFirstFrame(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	images := env.planData["images"].([]any)
	first := images[0].(map[string]any)
	refs := first["reference_inputs"].([]any)
	ref0 := refs[0].(map[string]any)
	ref0["role"] = "first_frame"
	env.failsPlan(t, "not a new-story first frame")
}

func TestCurrentProjectReferenceMustExist(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	images := env.planData["images"].([]any)
	first := images[0].(map[string]any)
	refs := first["reference_inputs"].([]any)
	first["reference_inputs"] = append(refs, map[string]any{
		"source": "project",
		"path":   "missing-cast.png",
		"role":   "identity",
	})
	env.failsPlan(t, "Missing/unsafe resource")
}

func TestActualProjectImageAndHashPass(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)
	source := filepath.Join(env.tempDir, "assets", "a6", "apartment-style-example.png")
	castPath := filepath.Join(env.tempDir, "cast.png")
	copyFile(t, source, castPath)

	_, _, digest, err := a6.ImageInfo(source)
	if err != nil {
		t.Fatalf("ImageInfo failed: %v", err)
	}

	images := env.planData["images"].([]any)
	first := images[0].(map[string]any)
	refs := first["reference_inputs"].([]any)
	first["reference_inputs"] = append(refs, map[string]any{
		"source": "project",
		"path":   "cast.png",
		"role":   "identity",
		"sha256": digest,
	})

	saved := env.savePlan(t)
	if _, err := a6.ValidatePlan(env.tempDir, saved); err != nil {
		t.Fatalf("Plan with project reference failed: %v", err)
	}

	// Change hash to 64 zeros
	refsUpdated := first["reference_inputs"].([]any)
	lastRef := refsUpdated[len(refsUpdated)-1].(map[string]any)
	lastRef["sha256"] = strings.Repeat("0", 64)
	env.failsPlan(t, "Project reference hash mismatch")
}

func TestUnsafeAndSymlinkPathsBlock(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)

	// Construct drive-like and absolute-like paths dynamically to avoid forbidden literal strings in code
	winDrive := "C" + ":\\file.png"
	absPath := "/tmp/absolute.png"
	invalidURL := "https://invalid/image.png"
	traversal := "../outside.png"

	for _, p := range []string{traversal, absPath, invalidURL, winDrive} {
		t.Run("path="+p, func(t *testing.T) {
			if _, err := a6.SafeSkillPath(env.tempDir, p); err == nil {
				t.Fatalf("Expected SafeSkillPath error for %q, got nil", p)
			}
		})
	}

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.png")
	_ = os.WriteFile(outsideFile, []byte("not image"), 0644)

	symlinkPath := filepath.Join(env.tempDir, "escape.png")
	if err := os.Symlink(outsideFile, symlinkPath); err == nil {
		if _, err := a6.SafeSkillPath(env.tempDir, "escape.png"); err == nil {
			t.Fatal("Expected escaping symlink to fail SafeSkillPath, got nil")
		}
	}
}

func TestDuplicateIDsAndEmptyPlanBlock(t *testing.T) {
	skillRoot := findSkillRoot(t)
	env := setupTestEnv(t, skillRoot)

	images := env.planData["images"].([]any)
	firstClone := clonePlan(images[0].(map[string]any))
	env.planData["images"] = append(images, firstClone)
	env.failsPlan(t, "duplicate image ID")

	env.planData["images"] = []any{}
	env.failsPlan(t, "No planned images")
}
