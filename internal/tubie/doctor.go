package tubie

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CheckTool verifies if an executable is found and runnable.
func CheckTool(name string) (bool, string) {
	path := BinaryPath(name)
	cmd := exec.Command(path, "-version")
	if name == "node" {
		cmd = exec.Command(path, "-v")
	}
	if err := cmd.Run(); err == nil {
		return true, path
	}
	// Try executing --help
	cmd = exec.Command(path, "--help")
	if err := cmd.Run(); err == nil {
		return true, path
	}
	return false, "PATH üzerinde bulunamadı"
}

// Doctor inspects system dependencies required for Tubie Studio.
func Doctor(root string) int {
	fmt.Println("=================================================================")
	fmt.Println("🩺 TUBIE ORTAM KONTROLÜ (DOCTOR)")
	fmt.Println("=================================================================")

	hasError := false

	// Required system binaries
	for _, bin := range []string{"node", "ffmpeg", "ffprobe"} {
		ok, path := CheckTool(bin)
		if ok {
			fmt.Printf("✅ OK %s: %s\n", bin, path)
		} else {
			fmt.Printf("❌ MISSING %s: %s\n", bin, path)
			hasError = true
		}
	}

	// Required Node canvas/sharp modules
	skillDir := filepath.Join(root, "skills", "directing-stickman-videos")
	for _, mod := range []string{"@napi-rs/canvas", "sharp"} {
		skillMod := filepath.Join(skillDir, "node_modules", mod)
		rootMod := filepath.Join(root, "node_modules", mod)
		if fi, err := os.Stat(skillMod); (err == nil && fi.IsDir()) || isDir(rootMod) {
			fmt.Printf("✅ OK node module %s\n", mod)
		} else {
			fmt.Printf("❌ MISSING node module %s: npm ci çalıştırın\n", mod)
			hasError = true
		}
	}

	// Edge-TTS for pro mode
	if _, err := exec.LookPath("edge-tts"); err == nil {
		fmt.Println("✅ OK edge-tts: Hazır (Profesyonel Türkçe seslendirme)")
	} else {
		// Check python -m edge_tts
		cmd := exec.Command(BinaryPath("python"), "-m", "edge_tts", "--help")
		if err := cmd.Run(); err == nil {
			fmt.Println("✅ OK edge-tts (Python): Hazır")
		} else {
			fmt.Println("⚠️  OPTIONAL edge-tts: pip install edge-tts")
		}
	}

	// Reference voice
	sesimWav := filepath.Join(root, "sesim.wav")
	if fi, err := os.Stat(sesimWav); err == nil && !fi.IsDir() && fi.Size() > 0 {
		fmt.Printf("✅ OK sesim.wav: Kayıtlı referans ses mevcut (%.1f KB)\n", float64(fi.Size())/1024.0)
	} else {
		fmt.Println("ℹ️  INFO sesim.wav: Henüz referans ses kaydedilmedi (GUI Ses sekmesinden ekleyebilirsiniz)")
	}

	if hasError {
		fmt.Println("\n❌ Bazı temel bağımlılıklar eksik. Lütfen eksikleri giderin.")
		return 1
	}

	fmt.Println("\n✨ Tüm temel ortam bağımlılıkları eksiksiz ve kullanıma hazır!")
	return 0
}

func isDir(p string) bool {
	if fi, err := os.Stat(p); err == nil && fi.IsDir() {
		return true
	}
	return false
}

