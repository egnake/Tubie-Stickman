# Kurulum ve sorun giderme

Ana kurulum komutları [README](../README.md) içindedir. Uygulama, bu deponun düzenlenebilir kurulumu için tasarlanmıştır; `pip install -e .` Python paketini kaynağa bağlar. Salt Python wheel dosyası, Node motorunu ve medya varlıklarını içermez.

## Python ve Node

Python 3.11 ortak başlangıç sürümüdür. Arch/CachyOS gibi sistem Python'u sık güncellenen dağıtımlarda `uv venv --python 3.11` ayrı bir ortam oluşturur; sistem paketlerini değiştirmez. Bağımlılıkları `sudo pip` ile kurmayın. [uv ortam belgeleri](https://docs.astral.sh/uv/pip/environments/).

Node motorunun bağımlılıkları kökteki npm workspace üzerinden `npm ci` ile kurulur. Kökteki `package-lock.json` tüm depo için kullanılır; skill altındaki kilit dosyası, skill tek başına dağıtıldığında kullanılır.

`doctor` komutunun çıkış kodu gerekli bileşen eksikse `1`, kontroller başarılıysa `0` olur. İsteğe bağlı mikrofon/XTTS paketleri ayrıca gösterilir. Bu kontrol gerçek mikrofon erişimini, model indirmeyi veya CUDA çalışma zamanını doğrulamaz.

## CPU ve CUDA

Temel profesyonel seslendirme yolu PyTorch gerektirmez. Faster-Whisper CPU üzerinde `int8` ile çalışabilir. XTTS için PyTorch gerekir; CUDA bulunamazsa CPU seçilir.

Yalnızca CPU ile XTTS kullanmak istiyorsanız, Linux ve Windows için önce PyTorch CPU paketlerini kurabilirsiniz:

```bash
python -m pip install torch torchaudio --index-url https://download.pytorch.org/whl/cpu
python -m pip install -e ".[clone]"
```

NVIDIA için sabit bir CUDA sürümü varsayılmaz. Sürücünüze uygun komutu [PyTorch seçicisinden](https://pytorch.org/get-started/locally/) alın. Klonlama ek paketlerini aynı ortamda kurun. Torch, torchaudio ve torchcodec sürümleri birlikte uyumlu olmalıdır; [Coqui kurulumu](https://coqui-tts.readthedocs.io/en/latest/installation.html) ve [TorchCodec uyumluluğu](https://github.com/pytorch/torchcodec#installing-torchcodec) esas alınmalıdır.

Faster-Whisper ayrı olarak CTranslate2 kullanır. CUDA'da sorun yaşanırsa [Faster-Whisper GPU gereksinimlerini](https://github.com/SYSTRAN/faster-whisper#gpu) kontrol edin. `torch.cuda.is_available()` sonucunun olumlu olması, cuDNN/CTranslate2 kitaplıklarının da hazır olduğu anlamına gelmez.

## FFmpeg, ses ve fontlar

`ffmpeg` ve `ffprobe` PATH üzerinde olmalıdır. Gerekirse `FFMPEG` ve `FFPROBE` ortam değişkenlerine yürütülebilir dosyanın tam yolunu verin. `NODE` değişkeni Python üretim akışının Node yolunu değiştirebilir. Yolun etrafına değişken değerinin bir parçası olarak tırnak eklemeyin.

GUI, tarayıcının mikrofonunu kullanır; Python `record` ek paketleri gerektirmez. Tarayıcı mikrofonu desteklemiyorsa Ses sekmesinden dosya yükleyin. Eski Python ses yardımcılarını doğrudan kullanmak için `.[record]` kurulabilir.

Arch/CachyOS'ta Python mikrofon yardımcıları için `portaudio` paketi ve çalışan bir ses oturumu gerekir. `sounddevice` ses aygıtı bulamıyorsa varsayılan giriş aygıtını sistem ayarlarından seçin. Windows'ta masaüstü uygulamalarının mikrofon iznini kontrol edin. Dosyadan ses yükleme seçeneği mikrofon gerektirmez.

A6 son işleme ve Node render motoru depodaki fontları kullanır; Windows veya macOS font klasörü gerekmez. Yardımcı görsel üretim betiği için `.[dev]` içindeki Pillow gerekir.

## Çevrimdışı kullanım

Profesyonel seslendirme (`pro`) Edge TTS hizmetine bağlanır. XTTS ve Whisper modelleri ilk kullanımda indirilir; modellerin önceden indirilmiş olması çevrimdışı üretim için gereklidir. Node render testleri sentetik yerel medya kullanır. Gelişmiş skill stillerinin bazıları ayrıca yapılandırılan harici görsel/ses sağlayıcılarını kullanır; ayrıntılar stil kataloğundadır.
