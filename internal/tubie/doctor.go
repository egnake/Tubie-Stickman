package tubie

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CheckTool verifies if an executable is found and runnable.
func CheckTool(name string) (bool, string) {
	path := BinaryPath(name)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "-version")
	if name == "node" {
		cmd = exec.CommandContext(ctx, path, "-v")
	}
	if err := cmd.Run(); err == nil {
		return true, path
	}
	// Try executing --help
	ctxHelp, cancelHelp := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelHelp()
	cmd = exec.CommandContext(ctxHelp, path, "--help")
	if err := cmd.Run(); err == nil {
		return true, path
	}
	return false, "PATH üzerinde bulunamadı"
}

func checkPythonCommand(args ...string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ResolvePythonBinary(), args...)
	cmd.Dir = os.TempDir()
	return cmd.Run() == nil
}

// DoctorWithLogger inspects system dependencies required for Tubie Studio and streams each line to logf.
func DoctorWithLogger(root string, logf func(string)) int {
	log := func(format string, args ...any) {
		if logf != nil {
			logf(fmt.Sprintf(format, args...))
		}
	}

	log("=================================================================")
	log("🩺 TUBIE ORTAM KONTROLÜ (DOCTOR)")
	log("=================================================================")

	hasError := false

	// Required system binaries
	for _, bin := range []string{"node", "ffmpeg", "ffprobe"} {
		ok, path := CheckTool(bin)
		if ok {
			log("✅ OK %s: %s", bin, path)
		} else {
			log("❌ MISSING %s: %s", bin, path)
			hasError = true
		}
	}

	// Required Node canvas/sharp modules
	skillDir := filepath.Join(root, "skills", "directing-stickman-videos")
	for _, mod := range []string{"@napi-rs/canvas", "sharp"} {
		skillMod := filepath.Join(skillDir, "node_modules", mod)
		rootMod := filepath.Join(root, "node_modules", mod)
		if fi, err := os.Stat(skillMod); (err == nil && fi.IsDir()) || isDir(rootMod) {
			log("✅ OK node module %s", mod)
		} else {
			log("❌ MISSING node module %s: npm ci çalıştırın", mod)
			hasError = true
		}
	}

	// Edge-TTS for pro mode
	if _, err := exec.LookPath("edge-tts"); err == nil {
		log("✅ OK edge-tts: Hazır (Profesyonel Türkçe seslendirme)")
	} else if checkPythonCommand("-m", "edge_tts", "--help") {
		log("✅ OK edge-tts (Python): Hazır")
	} else {
		log("⚠️  OPTIONAL edge-tts: pip install edge-tts")
	}

	// g4f for free cloud AI script and aciklama.txt generation
	if CheckG4FAvailable() {
		log("✅ OK g4f (Python): Hazır (Bulut yapay zeka senaryo ve açıklama üretimi)")
	} else {
		log("ℹ️  OPTIONAL g4f: pip install -U g4f (Bulut yapay zeka senaryo ve açıklama üretimi için)")
	}

	// Optional faster-whisper acoustic aligner
	if checkPythonCommand("-c", "import faster_whisper") {
		log("✅ OK faster-whisper (Python): Hazır (Milisaniye hassasiyetinde kelime hizalaması)")
	} else {
		log("ℹ️  INFO faster-whisper: Kurulu değil (opsiyonel; ses zarfı hizalaması kullanılacak)")
	}

	// Optional XTTS voice cloning
	if checkXTTSAvailable() {
		log("✅ OK TTS/XTTS (Python): Hazır (Klonlanmış ses sentezi · cihaz=%s)", ResolveComputeDevice())
	} else {
		log("ℹ️  INFO TTS/XTTS: Kurulu değil (opsiyonel; yalnızca 'Kendi sesim' modu için gerekli)")
	}

	// Reference voice
	sesimWav := filepath.Join(root, "sesim.wav")
	if fi, err := os.Stat(sesimWav); err == nil && !fi.IsDir() && fi.Size() > 0 {
		log("✅ OK sesim.wav: Kayıtlı referans ses mevcut (%.1f KB)", float64(fi.Size())/1024.0)
	} else {
		log("ℹ️  INFO sesim.wav: Henüz referans ses kaydedilmedi (GUI Ses sekmesinden ekleyebilirsiniz)")
	}

	if hasError {
		log("❌ Bazı temel bağımlılıklar eksik. Lütfen eksikleri giderin.")
		return 1
	}

	log("✨ Tüm temel ortam bağımlılıkları eksiksiz ve kullanıma hazır!")
	return 0
}

// Doctor inspects system dependencies required for Tubie Studio.
func Doctor(root string) int {
	return DoctorWithLogger(root, func(line string) {
		fmt.Println(line)
	})
}

func isDir(p string) bool {
	if fi, err := os.Stat(p); err == nil && fi.IsDir() {
		return true
	}
	return false
}
