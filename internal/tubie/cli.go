package tubie

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Main is the primary entrypoint for the Tubie CLI and Studio.
func Main(args []string) int {
	root, err := FindRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
		return 1
	}

	if len(args) == 0 {
		return runGUI(root, 0, true, "produce")
	}

	switch args[0] {
	case "gui":
		guiFlags := flag.NewFlagSet("gui", flag.ExitOnError)
		port := guiFlags.Int("port", 0, "Yerel port")
		noBrowser := guiFlags.Bool("no-browser", false, "Tarayıcıyı otomatik açma")
		_ = guiFlags.Parse(args[1:])
		return runGUI(root, *port, !*noBrowser, "produce")

	case "record":
		return runGUI(root, 0, true, "voice")

	case "doctor":
		return Doctor(root)

	case "generate":
		genFlags := flag.NewFlagSet("generate", flag.ExitOnError)
		preset := genFlags.String("preset", "", "Şablon numarası (1..10)")
		topic := genFlags.String("topic", "", "Özel konu / fikir (örn: 'Sahte linkleri 5 saniyede nasıl anlarsın?')")
		style := genFlags.String("style", StyleHybrid, "Senaryo tarzı: hybrid, cinematic, informative, mythbuster, challenge, redvsblue")
		all := genFlags.Bool("all", false, "Tüm şablonları üret (1..10)")
		mode := genFlags.String("mode", "auto", "Ses modu: pro, cloned, auto")
		voice := genFlags.String("voice", "viral-male", "Ses stili: viral-male, viral-fast, viral-female, deep-male, cloned, cloned-fast")
		bgm := genFlags.String("bgm", "cyber", "Arka plan müziği: cyber, thriller, lofi, none")
		format := genFlags.String("format", "shorts", "Video formatı: shorts (9:16 720p), shorts-hd (1080p), youtube (16:9 720p), youtube-hd (1080p)")
		loop := genFlags.Bool("loop", false, "Sonsuz döngü (loop) kapanışı ekle")
		_ = genFlags.Parse(args[1:])

		pipeline := &Pipeline{
			Root:       root,
			Mode:       *mode,
			VoiceStyle: *voice,
			BGM:        *bgm,
			Format:     *format,
			LoopEnding: *loop,
		}

		if *topic != "" {
			normStyle := NormalizeScriptStyle(*style)
			fmt.Printf("✨ Yapay zeka ile senaryo üretiliyor [%s, loop=%v]: %q...\n", normStyle, *loop, *topic)
			res, err := GenerateScriptWithOptions(*topic, ScriptOptions{
				Style:      normStyle,
				LoopEnding: *loop,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Senaryo üretim hatası: %v\n", err)
				return 1
			}
			customPreset := Preset{
				ID:        "custom",
				Title:     res.Title,
				Slug:      res.Slug,
				Narration: res.Narration,
				Scenes:    res.Scenes,
			}
			if _, err := pipeline.Produce(customPreset); err != nil {
				fmt.Fprintf(os.Stderr, "Üretim hatası (%s): %v\n", customPreset.Title, err)
				return 1
			}
			return 0
		}

		targets := []string{*preset}
		if *all {
			for _, item := range PresetList() {
				targets = append(targets, item.ID)
			}
			if len(targets) > 1 && targets[0] == "" {
				targets = targets[1:]
			}
		} else if *preset == "" {
			targets = []string{"1"}
		}

		for _, pKey := range targets {
			p, ok := CyberPresets[pKey]
			if !ok {
				fmt.Fprintf(os.Stderr, "Geçersiz şablon: %s (1..10 arası seçin)\n", pKey)
				return 1
			}
			if _, err := pipeline.Produce(p); err != nil {
				fmt.Fprintf(os.Stderr, "Üretim hatası (%s): %v\n", p.Title, err)
				return 1
			}
		}
		return 0

	case "describe":
		return runDescribe(args[1:])

	case "analyze":
		return runAnalyze(args[1:])

	case "--help", "-h", "help":
		printHelp()
		return 0

	default:
		fmt.Fprintf(os.Stderr, "Bilinmeyen komut: %s\n\n", args[0])
		printHelp()
		return 1
	}
}

func printHelp() {
	fmt.Println(`Tubie Studio - Siber Güvenlik Viral Video Motoru

Kullanım:
  go run .                          # Tubie Studio GUI (varsayılan)
  go run . gui [--port N] [--no-browser] # Arayüzü başlat
  go run . record                   # GUI ses merkezi sekmesini aç
  go run . doctor                   # Ortam ve bağımlılık kontrolü
  go run . generate --topic "İnstagram" [--style hybrid|cinematic|informative|mythbuster|challenge|redvsblue] [--format shorts|shorts-hd|youtube|youtube-hd] [--loop] [--bgm cyber|thriller|lofi|none]
  go run . generate --preset 1      # Hazır şablon üret (1..10)
  go run . generate --all           # Tüm 10 şablonu sırayla üret
  go run . generate --mode pro      # Edge-TTS spiker sesi ile üret
  go run . describe --title "..." --text "..." [--out aciklama.txt] # İçerikle uyumlu aciklama.txt üret
  go run . analyze --title "..." --text "..." [--loop] [--json]      # Viral tutma ve kanca skorunu analiz et`)
}

// runDescribe generates content-matched aciklama.txt metadata from a title and narration text.
func runDescribe(args []string) int {
	fs := flag.NewFlagSet("describe", flag.ContinueOnError)
	title := fs.String("title", "", "Video başlığı")
	text := fs.String("text", "", "Video seslendirme metni")
	out := fs.String("out", "", "Çıktı dosyası (boş bırakılırsa standart çıktıya yazar)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	t := strings.TrimSpace(*title)
	n := strings.TrimSpace(*text)
	if t == "" && n == "" {
		fmt.Fprintln(os.Stderr, "Hata: --text veya --title parametresi gerekli")
		return 1
	}

	vm := GenerateViralMetadata(t, n)
	content := FormatViralText(t, n, vm)

	outPath := strings.TrimSpace(*out)
	if outPath == "" {
		fmt.Print(content)
		return 0
	}

	cleanOut := filepath.Clean(outPath)
	if dir := filepath.Dir(cleanOut); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Klasör oluşturulamadı (%s): %v\n", dir, err)
			return 1
		}
	}
	if err := os.WriteFile(cleanOut, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Dosya yazılamadı (%s): %v\n", cleanOut, err)
		return 1
	}
	fmt.Printf("📝 Viral paylaşım paketi yazıldı: %s\n", cleanOut)
	return 0
}

// runAnalyze scores a title and narration script for cybersecurity viral retention.
func runAnalyze(args []string) int {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	title := fs.String("title", "", "Video başlığı")
	text := fs.String("text", "", "Video seslendirme metni")
	loop := fs.Bool("loop", false, "Sonsuz döngü kapanışı kontrolü")
	asJSON := fs.Bool("json", false, "Sonucu JSON formatında yazdır")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	t := strings.TrimSpace(*title)
	n := strings.TrimSpace(*text)
	if t == "" && n == "" {
		fmt.Fprintln(os.Stderr, "Hata: --text veya --title parametresi gerekli")
		return 1
	}

	scenes := EnsureScenesForScript(t, n)
	report := AnalyzeScriptVirality(t, n, scenes, *loop)
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(report)
		return 0
	}

	fmt.Printf("🔥 VİRAL SKOR ANALİZİ: %d/100 (Derece: %s) · %s\n", report.TotalScore, report.Grade, report.ThreatCategory)
	fmt.Printf("⏱️ Tahmini Süre: %ds (%d kelime) | Loop: %v\n", report.EstimatedSeconds, report.WordCount, report.LoopReady)
	if report.HookSentence != "" {
		fmt.Printf("🪝 Kanca: %s\n", report.HookSentence)
	}
	fmt.Println("-----------------------------------------------------------------")
	for _, dim := range report.Dimensions {
		fmt.Printf("• %-24s %2d/%2d  (%s)\n", dim.Name, dim.Score, dim.MaxScore, dim.Feedback)
	}
	if len(report.Tips) > 0 {
		fmt.Println("\n💡 İyileştirme Önerileri:")
		for _, tip := range report.Tips {
			fmt.Printf("  - %s\n", tip)
		}
	}
	return 0
}

func runGUI(root string, port int, openBrowser bool, tab string) int {
	server, err := NewStudioServer(root, port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Sunucu başlatılamadı: %v\n", err)
		return 1
	}

	url := server.URL(tab)
	fmt.Printf("\n🎬 Tubie Studio: %s\n", url)
	fmt.Println("Kapatmak için Ctrl+C veya GUI'deki 'Uygulamayı kapat' düğmesini kullanın.")

	if openBrowser {
		OpenBrowser(url)
	}

	if err := server.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "Sunucu kapandı: %v\n", err)
	}
	return 0
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

func init() {
	_ = filepath.Separator
}
