# Tubie-Stickman

Türkçe siber güvenlik videoları için seslendirme, kelime zamanlaması ve çöp adam animasyonu üretim araçları. Windows ve Linux için ortak Python komut arayüzü ve Node.js render motoru içerir.

## Kurulum

**Önerilen ortam:** Python 3.11, Node.js 22 LTS veya üzeri, FFmpeg ve FFprobe. Python komutları proje kökünden, sanal ortam içinde çalıştırılır. Python paketinin dışındaki görsel ve font dosyaları da kullanıldığı için **düzenlenebilir kurulum** (`-e`) gereklidir; depoyu kurulumdan sonra taşımayın.

### CachyOS / Arch Linux

```bash
sudo pacman -Syu --needed git nodejs-lts-jod npm ffmpeg portaudio uv

git clone https://github.com/egnake/Tubie-Stickman.git
cd Tubie-Stickman
uv venv --python 3.11
source .venv/bin/activate
uv pip install -e ".[record,dev]"
npm ci
./uret.sh doctor
./uret.sh
```

Arch'ın sistem Python sürümünü değiştirmeyin. `uv`, seçilen Python sürümünü ayrı olarak yönetir. Sistemde güncel Node.js zaten varsa ayrıca `nodejs-lts-jod` kurmanız gerekmez. Paket seçimi ve Python ortamları için [Arch paketleri](https://archlinux.org/packages/) ve [uv belgeleri](https://docs.astral.sh/uv/guides/install-python/).

### Diğer Linux dağıtımları

Dağıtımınızın paket yöneticisinden Node.js, FFmpeg ve mikrofon için PortAudio kurun. Ardından yukarıdaki `uv venv` ile başlayan adımları uygulayın. Uygun Python zaten kuruluysa `python3.11 -m venv .venv` ve etkinleştirmeden sonra `python -m pip install -e ".[record,dev]"` kullanılabilir.

### Windows 10 / 11

Python 3.11, Node.js ve FFmpeg kurun; `python`, `node`, `ffmpeg` ve `ffprobe` komutlarının terminalden bulunduğunu doğrulayın. Proje kökünde PowerShell veya CMD açın:

```powershell
py -3.11 -m venv .venv
.venv\Scripts\python.exe -m pip install -e ".[record,dev]"
npm ci
.\uret.bat doctor
.\uret.bat
```

Başlatıcılar projenin `.venv` ortamını otomatik seçer. Sanal ortamı elle etkinleştirmek zorunlu değildir.

## Kullanım

```bash
tubie                             # Etkileşimli menü
tubie doctor                      # Eksik bağımlılıkları göster
tubie generate --preset 1 --mode pro
tubie generate --all --mode pro
tubie record                      # Mikrofon veya mevcut dosyadan ses kaydı
tubie generate --preset 2 --mode cloned
```

`tubie` için sanal ortam etkin olmalıdır. Aynı seçenekler Linux'ta `./uret.sh`, Windows'ta `uret.bat` ile kullanılabilir. Ses kaydı için `sesimi_kaydet.sh` / `sesimi_kaydet.bat` da vardır.

| Şablon | Konu |
| --- | --- |
| 1 | Wi-Fi güvenliği |
| 2 | Brute force saldırıları |
| 3 | SQL injection |
| 4 | Phishing / oltalama |

`--mode pro` Edge TTS üzerinden ses üretir ve **internet gerektirir**. `--mode cloned` yerel XTTS-v2 kullanır; ek kurulum ve `sesim.wav` referansı gerekir. `auto`, referans varsa klonlama yolunu seçer. Modeller ilk kullanımda indirilir. CUDA bulunmadığında CPU kullanılır; CPU ile üretim daha uzun sürer.

Üretilen videolar `output/`, ara dosyalar `projects/` altına yazılır. Bunlar ve kişisel `sesim.wav` kaydı Git tarafından dışlanır. Depodaki mevcut örnekler `examples/` altındadır.

## Ses klonlama ve GPU

```bash
# Etkin sanal ortamda, isteğe bağlı XTTS bağımlılıkları:
python -m pip install -e ".[clone]"
```

GPU kullanımı için donanımınıza ve işletim sisteminize uygun PyTorch kurulumunu [resmî seçiciden](https://pytorch.org/get-started/locally/) yapın. PyTorch CUDA erişimi ile Faster-Whisper'ın CUDA/cuDNN gereksinimleri ayrıdır; ayrıntılar [kurulum rehberinde](docs/setup.md). CPU kullanımı için NVIDIA kart gerekmez. XTTS modelinin lisans koşulları ayrıca değerlendirilmelidir.

## Klasör yapısı

```text
src/tubie_stickman/             Python uygulaması ve ortak CLI
skills/directing-stickman-videos/  Node render motoru, stil varlıkları ve skill
assets/                        Ortak ses dosyaları ve katalog medyası
docs/                          Kurulum, kullanım, stil kataloğu ve QA kayıtları
examples/                      Korunan örnek proje ve video çıktıları
scripts/                       Geliştirme ve varlık üretim araçları
tests/                         Python testleri ve belge/kontrat kontrolleri
projects/                      Üretim sırasında oluşan ara dosyalar (Git dışı)
output/                        Yeni video çıktıları (Git dışı)
```

Render motoru kendi içinde dağıtılabilen bir skill olduğu için bağımlılık ve varlıkları `skills/` altında birlikte tutulur.

## Geliştirme

```bash
npm ci
python -m pip install -e ".[dev]"
python scripts/check.py
```

Kontroller Python testlerini, belgeleri ve Node render testlerini çalıştırır; ücretli API çağrısı veya model indirmesi yapmaz. CI aynı kontrolleri Windows ve Linux üzerinde çalıştırmak üzere yapılandırılmıştır.

- [Kurulum ve sorun giderme](docs/setup.md)
- [Kullanım rehberi](docs/usage.md)
- [Türkçe stil kataloğu](docs/catalog/README.tr.md) · [English style catalog](docs/catalog/README.en.md)
- [Örnek dosyalar](examples/README.md)

[MIT lisansı](LICENSE)
