# Tubie-Stickman

Türkçe siber güvenlik videoları için seslendirme, kelime zamanlaması ve çöp adam animasyonu üretim araçları. Windows ve Linux için koyu temalı Tubie Studio GUI’si ve Node.js render motoru içerir.

## Kurulum

**Önerilen ortam:** Python 3.11, Node.js 22 LTS veya üzeri, FFmpeg ve FFprobe. Python komutları proje kökünden, sanal ortam içinde çalıştırılır. Python paketinin dışındaki görsel ve font dosyaları da kullanıldığı için **düzenlenebilir kurulum** (`-e`) gereklidir; depoyu kurulumdan sonra taşımayın.

### CachyOS / Arch Linux

```bash
sudo pacman -Syu --needed git nodejs-lts-jod npm ffmpeg portaudio uv

git clone https://github.com/egnake/Tubie-Stickman.git
cd Tubie-Stickman
uv venv --python 3.11
source .venv/bin/activate
uv pip install -e ".[dev]"
npm ci
python baslat.py
```

Arch'ın sistem Python sürümünü değiştirmeyin. `uv`, seçilen Python sürümünü ayrı olarak yönetir. Sistemde güncel Node.js zaten varsa ayrıca `nodejs-lts-jod` kurmanız gerekmez. Paket seçimi ve Python ortamları için [Arch paketleri](https://archlinux.org/packages/) ve [uv belgeleri](https://docs.astral.sh/uv/guides/install-python/).

### Diğer Linux dağıtımları

Dağıtımınızın paket yöneticisinden Node.js, FFmpeg ve mikrofon için PortAudio kurun. Ardından yukarıdaki `uv venv` ile başlayan adımları uygulayın. Uygun Python zaten kuruluysa `python3.11 -m venv .venv` ve etkinleştirmeden sonra `python -m pip install -e ".[dev]"` kullanılabilir.

### Windows 10 / 11

Python 3.11, Node.js ve FFmpeg kurun; `python`, `node`, `ffmpeg` ve `ffprobe` komutlarının terminalden bulunduğunu doğrulayın. Proje kökünde PowerShell veya CMD açın:

```powershell
py -3.11 -m venv .venv
.venv\Scripts\python.exe -m pip install -e ".[dev]"
npm ci
py baslat.py
```

Tek başlatıcı `baslat.py`, projenin `.venv` ortamını otomatik seçer ve Tubie Studio’yu varsayılan tarayıcıda açar. Sanal ortamı elle etkinleştirmek zorunlu değildir. Linux’ta `python3 baslat.py`, Windows’ta `py baslat.py` kullanın. Arayüz yalnızca bu bilgisayardan erişilen rastgele bir yerel adreste çalışır; ek bir GUI paketi gerekmez.

## Kullanım

GUI’den konu ve ses seçimi, sıralı üretim, referans ses yükleme, 15 saniyelik mikrofon kaydı, klonlama testi, ortam kontrolü ve çıktı indirme yönetilir. Üretim sırasında gerçek render kareleri, işlenen kare sayısı, aşamalar ve işlem günlüğü canlı görünür. Önizleme hazır olduğunda videoyu izleyip **Onayla ve tamamla** düğmesine basın. İşlemi GUI’den iptal edebilirsiniz.

Arayüzü kapatmak için **Uygulamayı kapat** düğmesini kullanın. Tarayıcı sekmesini kapatmak üretimi durdurmaz. Aynı bilgisayardaki başka bir tarayıcıdan, terminalde yazan adresle devam edebilirsiniz.

Geliştirme ve otomasyon için komut satırı da kullanılabilir:

```bash
python baslat.py                   # Tubie Studio GUI
python baslat.py record            # GUI ses merkezi
python baslat.py doctor            # Bağımlılık kontrolü
python baslat.py generate --preset 1 --mode pro
python baslat.py generate --all --mode pro
python baslat.py gui --no-browser  # Adresi yazdır; tarayıcıyı elle aç
```

Etkin sanal ortamda aynı komutlar `tubie` ile de kullanılabilir. `.bat` ve `.sh` başlatıcıları kaldırılmıştır.

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
src/tubie_stickman/             Python uygulaması, GUI ve CLI
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
