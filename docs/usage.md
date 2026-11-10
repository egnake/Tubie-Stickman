# Kullanım rehberi

Önce [kurulumu](../README.md#kurulum) tamamlayın. Linux'ta `./uret.sh`, Windows'ta `uret.bat` aynı menüyü açar. Komutları herhangi bir çalışma klasöründen, başlatıcının tam yoluyla da çalıştırabilirsiniz.

## Video üretimi

```bash
tubie generate --preset 1 --mode pro
tubie generate --all --mode pro
```

Şablonlar: `1` Wi-Fi, `2` brute force, `3` SQL injection, `4` phishing. `--all` dört şablonu sırayla üretir.

Profesyonel ses Edge TTS kullanır; internet gerekir. Kelime zamanlamaları Faster-Whisper ile çıkarılır. Son video, altyazı ve önizleme `output/` altına; çalışma dosyaları `projects/<şablon>/` altına kaydedilir. Aynı şablon yeniden üretildiğinde çalışma ve çıktı dosyaları yenilenir.

## Kendi sesinizi kullanma

Mikrofon kaydı için `.[record]`, klonlama için `.[clone]` ek paketlerini kurun. [Ek kurulum ayrıntıları](setup.md).

```bash
tubie record
```

Menüden 15 saniyelik mikrofon kaydını veya mevcut ses dosyasını seçin. Referans `sesim.wav` adıyla proje köküne kaydedilir. Bu kişisel dosya Git'e dahil edilmez. Yeni kayıt mevcut referansı değiştirir.

```bash
tubie generate --preset 1 --mode cloned
```

`cloned` modu referans yoksa hata verir. `auto`, `sesim.wav` varsa klonlama kullanır, yoksa profesyonel sese geçer. Etkileşimli menüde ses modu açıkça seçilir.

## Kontrol ve gelişmiş stiller

```bash
tubie doctor
npm run doctor
```

İlk komut Türkçe üretim uygulamasının bağımlılıklarını, ikinci komut gelişmiş Node motorunun ortamını kontrol eder. Node motorunun bazı sağlayıcı anahtarları Türkçe hazır şablonları çalıştırmak için gerekli değildir.

[Türkçe stil kataloğu](catalog/README.tr.md) ve [İngilizce katalog](catalog/README.en.md), skill üzerinden yönetilen Director/Autopilot akışlarını ve diğer stilleri açıklar. Bu akışlar dört hazır Türkçe CLI şablonundan daha geniştir.
