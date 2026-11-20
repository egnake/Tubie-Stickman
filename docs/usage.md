# Kullanım rehberi

Önce [kurulumu](../README.md#kurulum) tamamlayın. `go run .` komutu **Tubie Studio Pro** arayüzünü açar. İsteğe bağlı olarak derlenmiş `./tubie` (Windows'ta `.\tubie.exe`) ikilisi doğrudan çalıştırılabilir.

## GUI ile video üretimi

1. **Konu, Fikir & Senaryo (`01`)**:
   - **Yapay Zeka & Trendler** sekmesinde 18 kategorize siber güvenlik fikrinden birini seçebilir veya kendi konunuzu yazabilirsiniz.
   - 6 anlatım formatından (`hybrid`, `cinematic`, `informative`, `mythbuster`, `challenge`, `redvsblue`) birini seçin ve dilerseniz **Sonsuz Döngü (Loop) Modu** kutusunu işaretleyin.
   - **10 Hazır Şablon** sekmesinden 1..10 arası şablonları tek tek veya tümünü sıraya alarak üretebilirsiniz.
   - Senaryo metnini düzenledikçe **Viral Tutma Analizi** kartı 0–100 arası canlı skor (`S/A/B/C/D`), 6 alt boyut ve somut iyileştirme önerileri gösterir.
   - **Etkileşimli Sahne Yönetmeni (6 Sahne)** panelinden her sahnenin başlığını, duygu tonunu (`mood`), veri widget türünü (`visual_type`), kamera hareketini ve çağrı etiketini anlık özelleştirebilirsiniz.
   - **Kapak Önizle (9:16 + 16:9)** düğmesiyle üretime başlamadan önce yüksek tıklanma odaklı A/B ve yatay YouTube kapaklarını görebilirsiniz.

2. **Ses, Müzik & Format (`02`)**:
   - **Video Çözünürlük & Oranı**: `shorts` (9:16 · 720×1280), `shorts-hd` (9:16 · 1080×1920), `youtube` (16:9 · 1280×720) veya `youtube-hd` (16:9 · 1920×1080).
   - **Seslendirme Stili**: `viral-male`, `viral-fast`, `viral-female`, `deep-male` (Edge TTS + stüdyo EQ/kompresör/silence tightening) veya `cloned` / `cloned-fast` (yerel `sesim.wav`). Profesyonel ses Edge TTS internet gerektirir.
   - **Arka Plan Müziği**: `cyber`, `thriller`, `lofi` veya `none`. Konuşma sırasında müzik FFmpeg `sidechaincompress` ile otomatik dengelenir.

3. **Canlı Önizleme, Onay & Çıktı Paketi**:
   - Üretim sırasında gerçek render kareleri yaklaşık yarım saniye aralıklarla canlı görünür.
   - Önizleme oluşunca GUI’de oynatın ve **Onayla ve tamamla** ile son MP4 üretimine geçin. **İptal et**, üretimi alt süreçleriyle birlikte güvenle durdurur.
   - **Çıktılar** sekmesinde ve `output/<slug>/` klasöründe her video için `.mp4`, `.srt`, `.vtt`, `thumbnail.jpg` (9:16 A), `thumbnail_b.jpg` (9:16 B), `thumbnail_16x9.jpg` (16:9), `viral_report.json` ve tek tıkla kopyalanabilir `aciklama.txt` + sabit yorum saklanır.

## Kendi sesinizi kullanma

**Ses** sekmesinde mevcut ses dosyasını yükleyin ya da tarayıcı mikrofon izni vererek 15 saniyelik kayıt alın. Kaydı erken bitirmek de mümkündür. FFmpeg, sesi mono 24 kHz WAV biçimine dönüştürür. En fazla 50 MB kabul edilir. Yeni referans `sesim.wav` adıyla kaydedilir; önceki referans yalnızca dönüşüm başarılı olunca değiştirilir. Kayıt Git’e dahil edilmez.

Tarayıcı mikrofonu kullanıldığı için ek kurulum gerekmez; [kurulum ayrıntıları](setup.md). Referans olmadan kendi sesiniz modu başlatılmaz. Otomatik mod referans varsa kendi sesinizi, yoksa profesyonel sesi kullanır.

## Ortam kontrolü ve komut satırı

**Ortam** sekmesindeki kontrol, Node.js, FFmpeg, FFprobe, `@napi-rs/canvas`, `sharp`, `edge-tts`, `g4f`, isteğe bağlı `faster-whisper` ve `sesim.wav` durumunu günlükte listeler. Sunucu yalnızca `127.0.0.1` üzerinde oturuma özel rastgele bir token adresinde ve katı CSP + boyut sınırlarıyla çalışır.

```bash
go run . --help
go run . gui --no-browser
go run . gui --port 8765
go run . doctor
go run . generate --topic "Oltalama tuzakları" --style challenge --format shorts-hd --loop --bgm thriller --mode pro
go run . generate --preset 1 --format youtube --mode pro
go run . generate --all --mode pro
go run . analyze --title "Sahte Link" --text "Bu linke tıklama!" --loop --json
go run . describe --title "Sahte Link" --text "Bu linke tıklama!" --out aciklama.txt
npm run doctor
```

[Türkçe stil kataloğu](catalog/README.tr.md) ve [İngilizce katalog](catalog/README.en.md), skill üzerinden yönetilen Director/Autopilot akışlarını açıklar. `Tubie Studio` siber güvenlik kanalı için varsayılan Türkçe dikey ve yatay şablonları ve prosedürel görselleri yönetir; katalogdaki diğer stiller skill akışında kullanılır.
