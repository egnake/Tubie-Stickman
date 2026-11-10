package tubie

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
		preset := genFlags.String("preset", "", "Şablon numarası (1, 2, 3, 4)")
		topic := genFlags.String("topic", "", "Özel konu / fikir (örn: 'İnstagram hesap çalma')")
		all := genFlags.Bool("all", false, "Tüm şablonları üret")
		mode := genFlags.String("mode", "auto", "Ses modu: pro, cloned, auto")
		voice := genFlags.String("voice", "viral-male", "Ses stili: viral-male, viral-fast, viral-female, deep-male, cloned")
		bgm := genFlags.String("bgm", "cyber", "Arka plan müziği: cyber, lofi, none")
		_ = genFlags.Parse(args[1:])

		pipeline := &Pipeline{
			Root:       root,
			Mode:       *mode,
			VoiceStyle: *voice,
			BGM:        *bgm,
		}

		if *topic != "" {
			fmt.Printf("✨ Yapay zeka ile senaryo üretiliyor: %q...\n", *topic)
			res, err := GenerateScript(*topic)
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
			targets = []string{"1", "2", "3", "4"}
		} else if *preset == "" {
			targets = []string{"1"}
		}

		for _, pKey := range targets {
			p, ok := CyberPresets[pKey]
			if !ok {
				fmt.Fprintf(os.Stderr, "Geçersiz şablon: %s\n", pKey)
				return 1
			}
			if _, err := pipeline.Produce(p); err != nil {
				fmt.Fprintf(os.Stderr, "Üretim hatası (%s): %v\n", p.Title, err)
				return 1
			}
		}
		return 0

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
  go run . generate --topic "İnstagram" # Yapay zeka ile senaryo üretip video yap
  go run . generate --preset 1      # 1 numaralı hazır videoyu üret
  go run . generate --all           # Tüm şablonları sırayla üret
  go run . generate --mode pro      # Edge-TTS spiker sesi ile üret`)
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

