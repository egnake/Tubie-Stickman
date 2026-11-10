package tubie_test

import (
	"os"
	"path/filepath"
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

