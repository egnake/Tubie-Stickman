# Tubie-Stickman

Türkçe siber güvenlik videoları için seslendirme, kelime zamanlaması ve çöp adam animasyonu üretim araçları. Windows ve Linux için koyu temalı Tubie Studio GUI’si ve Go tabanlı video üretim motoru içerir.

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

Tubie Studio varsayılan tarayıcıda otomatik olarak açılır. Arayüz yalnızca bu bilgisayardan erişilen rastgele bir yerel adreste çalışır; ek bir GUI paketi gerekmez.

## Kullanım

GUI’den konu ve ses seçimi, sıralı üretim, referans ses yükleme, 15 saniyelik mikrofon kaydı, klonlama testi, ortam kontrolü ve çıktı indirme yönetilir. Üretim sırasında gerçek render kareleri, işlenen kare sayısı, aşamalar ve işlem günlüğü canlı görünür. Önizleme hazır olduğunda videoyu izleyip **Onayla ve tamamla** düğmesine basın. İşlemi GUI’den iptal edebilirsiniz.

Arayüzü kapatmak için **Uygulamayı kapat** düğmesini kullanın. Tarayıcı sekmesini kapatmak üretimi durdurmaz. Aynı bilgisayardaki başka bir tarayıcıdan, terminalde yazan adresle devam edebilirsiniz.

Geliştirme ve otomasyon için komut satırı da kullanılabilir:

```bash
go run .                          # Tubie Studio GUI
go run . record                   # GUI ses merkezi
go run . doctor                   # Bağımlılık kontrolü
go run . generate --preset 1 --mode pro
go run . generate --all --mode pro
go run . gui --no-browser         # Adresi yazdır; tarayıcıyı elle aç
```

Derlenmiş ikili ile aynı komutlar `./tubie` (veya Windows'ta `.\tubie.exe`) ile doğrudan çalıştırılabilir. `.bat`, `.sh` ve Python başlatıcıları kaldırılmıştır.

| Şablon | Konu |
| --- | --- |
| 1 | Wi-Fi güvenliği |
| 2 | Brute force saldırıları |
| 3 | SQL injection |
| 4 | Phishing / oltalama |

`--mode pro` Edge TTS üzerinden ses üretir ve **internet gerektirir**. `--mode cloned` yerel ses klonlama referansı (`sesim.wav`) kullanır. `auto`, referans varsa klonlama yolunu seçer.

Üretilen videolar `output/`, ara dosyalar `projects/` altına yazılır. Bunlar ve kişisel `sesim.wav` kaydı Git tarafından dışlanır. Depodaki mevcut örnekler `examples/` altındadır.

## Ses klonlama ve GPU

Temel profesyonel seslendirme modu (`pro`) doğrudan Edge TTS üzerinden çalışır ve harici Python veya GPU kütüphanelerine ihtiyaç duymaz. Kendi sesinizi kullanmak için GUI Ses sekmesinden 15 saniyelik referans kaydı alabilir veya hazır bir WAV dosyası yükleyebilirsiniz.

## Klasör yapısı

```text
cmd/                           Tubie CLI ve yardımcı araçlar
internal/tubie/                Go tabanlı video işleme, GUI ve CLI mantığı
skills/directing-stickman-videos/  Node render motoru, stil varlıkları ve skill
assets/                        Ortak ses dosyaları ve katalog medyası
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

Kontroller Go testlerini (`internal/tubie` ve `tests/verify_test.go`), belgeleri ve Node render testlerini çalıştırır; ücretli API çağrısı veya model indirmesi yapmaz. CI aynı kontrolleri Windows ve Linux üzerinde çalıştırmak üzere yapılandırılmıştır.

- [Kurulum ve sorun giderme](docs/setup.md)
- [Kullanım rehberi](docs/usage.md)
- [Türkçe stil kataloğu](docs/catalog/README.tr.md) · [English style catalog](docs/catalog/README.en.md)
- [Örnek dosyalar](examples/README.md)

[MIT lisansı](LICENSE)
