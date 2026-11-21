# Tubie-Stickman

Türkçe siber güvenlik videoları için viral kanca/tutma analizi, yapay zeka destekli senaryo üretimi, kelime zamanlaması, prosedürel çöp adam & hoodie'li hacker animasyonu, canlı 8 veri widget'ı ve ses mastering stüdyosu. Windows ve Linux için koyu temalı **Tubie Studio Pro** GUI’si ve Go tabanlı video üretim motoru içerir.

## Öne Çıkan Özellikler

- **6 Viral Siber Güvenlik Formatı**: `hybrid` (Hibrit Viral Akış), `cinematic` (Sinematik Gerilim), `informative` (3 Kural & Checklist), `mythbuster` (Efsane Yıkıcı), `challenge` (5 Saniye Testi), `redvsblue` (Saldırı vs Savunma).
- **18 Trend Konu Havuzu & 10 Hazır Şablon**: Oltalama, Yapay Zeka/Deepfake, Wi-Fi Evil Twin, Zararlı Yazılım/Infostealer, Mobil/Kripto ve Web/SQLi kategorilerinde hazır senaryolar.
- **Viral Tutma Analizörü (`analyze`)**: İlk 3 saniye kancası, gerilim yayı, tempo/ritim, teknik somutluk, kaydetme/paylaşma çağrısı ve döngü uyumu üzerinden **0–100 skor (`S/A/B/C/D`)** ve iyileştirme önerileri.
- **Sonsuz Döngü (Loop) Modu (`--loop`)**: Son cümleyi ilk açılış kancasına pürüzsüz bağlayarak tekrar izlenme (rewatch) oranını artırır.
- **Çift En-Boy Oranı (`9:16` + `16:9`)**: `shorts` (720×1280), `shorts-hd` (1080×1920), `youtube` (1280×720) ve `youtube-hd` (1920×1080) çıkış profilleri.
- **Prosedürel Siber Çöp Adam + 8 Canlı Widget**: Kurban/kahraman çöp adam, kırmızı vizörlü kapüşonlu hacker (`cyber_actor`), `notification`, `terminal`, `url_inspect`, `network_mitm`, `password_crack`, `threat_meter`, `checklist`, `versus` widget'ları, dinamik kamera hareketleri (`punch_in`, `zoom_slow`, `shake`, `glitch_cut`) ve Türkçe (`İ/Ş/Ğ`) uyumlu kelime-odaklı karaoke altyazılar.
- **Broadcast Ses & Sidechain Mix**: Sessizlik sıkıştırma (`silenceremove`), stüdyo EQ ve `-15 LUFS` normalizasyon; 8 prosedürel siber SFX ve konuşma sırasında arka plan müziğini (`cyber`, `thriller`, `lofi`) otomatik alçaltan `sidechaincompress`.
- **A/B Yüksek Tıklama Kapakları & Yayın Paketi**: Her video için `thumbnail.jpg` (9:16 A), `thumbnail_b.jpg` (9:16 B), `thumbnail_16x9.jpg` (16:9), `.srt`, `.vtt`, `viral_report.json` ve 5 A/B/C/D/E başlıklı `aciklama.txt` otomatik dışa aktarılır.

## Kurulum

**Önerilen ortam:** Go 1.22+, Node.js 22 LTS veya üzeri, FFmpeg ve FFprobe. Uygulama, proje kökünden Go araç zinciri ile doğrudan çalıştırılır veya derlenir. Paket dışındaki görsel ve font dosyaları da kullanıldığı için depoyu kurulumdan sonra taşımayın.

### CachyOS / Arch Linux

```bash
sudo pacman -Syu --needed git nodejs-lts-jod npm ffmpeg go

git clone https://github.com/egnake/Tubie-Stickman.git
cd Tubie-Stickman
npm ci
go run .
```

Sistemde güncel Node.js ve Go zaten varsa ek paket kurmanız gerekmez.

### Diğer Linux dağıtımları

Dağıtımınızın paket yöneticisinden Go 1.22+, Node.js ve FFmpeg kurun. Ardından proje kökünde `npm ci` ve `go run .` komutlarıyla uygulamayı başlatın.

### Windows 10 / 11

Go 1.22+, Node.js ve FFmpeg kurun; `go`, `node`, `ffmpeg` ve `ffprobe` komutlarının terminalden bulunduğunu doğrulayın. Proje kökünde PowerShell veya CMD açın:

```powershell
npm ci
go run .
```

İsteğe bağlı olarak tek bir yürütülebilir ikili derleyebilirsiniz:

```powershell
go build -o tubie.exe .
.\tubie.exe
```

Tubie Studio Pro varsayılan tarayıcıda otomatik olarak açılır. Arayüz yalnızca bu bilgisayardan erişilen rastgele bir yerel adreste çalışır; ek bir GUI paketi gerekmez.

## Kullanım

GUI’den yapay zeka senaryo üretimi, 18 trend fikir filtresi, canlı viral tutma skoru, anlık 9:16 + 16:9 kapak önizleme, etkileşimli 6 sahne yönetmeni, format/döngü seçimi, referans ses yükleme, 15 saniyelik mikrofon kaydı, klonlama testi, ortam kontrolü ve çıktı indirme yönetilir. Üretim sırasında gerçek render kareleri, işlenen kare sayısı, aşamalar ve işlem günlüğü canlı görünür. Önizleme hazır olduğunda videoyu izleyip **Onayla ve tamamla** düğmesine basın.

Arayüzü kapatmak için **Uygulamayı kapat** düğmesini kullanın. Tarayıcı sekmesini kapatmak üretimi durdurmaz. Aynı bilgisayardaki başka bir tarayıcıdan, terminalde yazan adresle devam edebilirsiniz.

Geliştirme ve otomasyon için komut satırı da kullanılabilir:

```bash
go run .                          # Tubie Studio Pro GUI
go run . record                   # GUI ses merkezi
go run . doctor                   # Ortam ve bağımlılık kontrolü (faster-whisper dahil)
go run . generate --topic "İnstagram" --style redvsblue --format shorts-hd --loop --bgm thriller --mode pro
go run . generate --preset 1 --format youtube --mode pro
go run . generate --all --mode pro
go run . analyze --title "Sahte Link" --text "Tıklama! 3 saniyede şifren çalınır." --loop
go run . describe --title "..." --text "..." --out aciklama.txt
go run . gui --no-browser         # Adresi yazdır; tarayıcıyı elle aç
```

Derlenmiş ikili ile aynı komutlar `./tubie` (veya Windows'ta `.\tubie.exe`) ile doğrudan çalıştırılabilir. `.bat`, `.sh` ve Python başlatıcıları kaldırılmıştır.

| Şablon | Kategori | Konu |
| --- | --- | --- |
| 1 | Ağ, Wi-Fi & Kablosuz | Wi-Fi Şifreniz 10 Saniyede Nasıl Kırılır? |
| 2 | Parola & Kimlik | Şifreniz Ne Kadar Sürede Kırılır? (Brute Force Gerçeği) |
| 3 | Web, Veri & Kod Güvenliği | SQL Injection Nedir ve Nasıl Çalışır? |
| 4 | Oltalama & Sosyal Mühendislik | Oltalama (Phishing) Tuzağına Nasıl Düşürülürsünüz? |
| 5 | Yapay Zeka & Deepfake | 3 Saniyelik Sesinle Nasıl Dolandırılırsın? |
| 6 | Kripto & Web3 Güvenliği | Kripto Cüzdanın Tek Onayla Nasıl Boşaltılır? |
| 7 | Zararlı Yazılım & Cihaz | Ücretsiz Crack İndirirken Bilgisayarında Ne Olur? |
| 8 | Mobil & Kripto | WhatsApp SMS Kodu & SIM Swap Hilesi |
| 9 | Yapay Zeka & Deepfake | Yapay Zeka Seni İnternette Ne Kadar Tanıyor? |
| 10 | Oltalama & Sosyal Mühendislik | Sokaktaki QR Kodu Okutmadan Önce Dur! |

`--mode pro` Edge TTS üzerinden ses üretir ve **internet gerektirir**. `--mode cloned` yerel ses klonlama referansı (`sesim.wav`) kullanır. `auto`, özel bir Pro ses stili seçilmediyse ve `sesim.wav` ile Python `TTS` (XTTS) hazırsa klonlama yolunu, aksi hâlde Edge TTS yolunu seçer.

Üretilen videolar `output/`, ara dosyalar `projects/` altına yazılır. Bunlar ve kişisel `sesim.wav` kaydı Git tarafından dışlanır. Depodaki mevcut örnekler `examples/` altındadır.

## Ses klonlama ve GPU

Temel profesyonel seslendirme modu (`pro`) doğrudan Edge TTS üzerinden çalışır ve harici Python veya GPU kütüphanelerine ihtiyaç duymaz. Kendi sesinizi kullanmak için GUI Ses sekmesinden 15 saniyelik referans kaydı alabilir veya hazır bir WAV dosyası yükleyebilirsiniz. İsteğe bağlı `faster-whisper` paketi kuruluysa kelime bazlı zamanlama milisaniye düzeyinde akustik doğrulamayla hizalanır; kurulu değilse çevrimdışı enerji zarfı hizalaması otomatik devreye girer.

## Klasör yapısı

```text
cmd/                           Tubie CLI ve test/doğrulama araçları
internal/tubie/                Go tabanlı viral analiz, senaryo, ses, kapak ve GUI motoru
skills/directing-stickman-videos/  Node render motoru, prosedürel siber görseller ve skill
assets/                        Arka plan müzikleri (cyber, thriller, lofi) ve katalog medyası
docs/                          Kurulum, kullanım, stil kataloğu ve QA kayıtları
examples/                      Korunan örnek proje ve video çıktıları
tests/                         Go regresyon testleri ve kontrat kontrolleri
projects/                      Üretim sırasında oluşan ara dosyalar (Git dışı)
output/                        Yeni video çıktıları (Git dışı)
```

Render motoru kendi içinde dağıtılabilen bir skill olduğu için bağımlılık ve varlıkları `skills/` altında birlikte tutulur.

## Geliştirme

```bash
npm ci
go run ./cmd/check
```

Kontroller Go testlerini (`internal/tubie` ve `tests/verify_test.go`), belgeleri ve Node render testlerini (`skills/directing-stickman-videos/tests/`) çevrimdışı olarak çalıştırır; ücretli API çağrısı veya model indirmesi yapmaz. CI aynı kontrolleri Windows ve Linux üzerinde çalıştırmak üzere yapılandırılmıştır.

- [Kurulum ve sorun giderme](docs/setup.md)
- [Kullanım rehberi](docs/usage.md)
- [Türkçe stil kataloğu](docs/catalog/README.tr.md) · [English style catalog](docs/catalog/README.en.md)
- [Örnek dosyalar](examples/README.md)

[MIT lisansı](LICENSE)
