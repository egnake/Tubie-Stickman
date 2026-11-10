# Kullanım rehberi

Önce [kurulumu](../README.md#kurulum) tamamlayın. Linux’ta `python3 baslat.py`, Windows’ta `py baslat.py` Tubie Studio’yu açar. Başlatıcı `.venv` ortamını kendisi seçer; tam yoluyla başka bir klasörden de çalıştırılabilir.

## GUI ile video üretimi

Üretim ekranında dört konudan birini seçin veya **Tümünü sıraya al** seçeneğini işaretleyin. Profesyonel, kendi sesiniz veya otomatik ses modunu seçerek **Üretime başla** düğmesine basın.

Profesyonel ses Edge TTS kullanır; internet gerekir. Kelime zamanlamaları Faster-Whisper ile çıkarılır. XTTS ve Whisper modelleri ilk kullanımda indirilir. GUI’de seslendirme, altyazı, önizleme, render ve çıktı aşamaları görünür. Gerçek render kareleri ve kare sayısı yaklaşık yarım saniye aralıklarla güncellenir; bu görüntü gerçek zamanlı video oynatımı değildir. Ses/model hazırlanırken ilk kare henüz oluşmaz; bu sırada günlük ve aşama bilgisi canlıdır.

Önizleme oluşunca GUI’de oynatın ve **Onayla ve tamamla** ile son MP4 üretimine geçin. Sıralı üretimde her videonun önizlemesi ayrı onaylanır. **İptal et**, üretim işlemini ve alt işlemlerini durdurur. Tamamlanmış çıktılar korunur; yarım kalan üretimler son video olarak gösterilmez. Yeni bir üretim başlatarak tekrar deneyebilirsiniz.

Son videolar ve altyazılar **Çıktılar** sekmesinden izlenir/indirilir. Dosyalar `output/`, çalışma dosyaları `projects/<şablon>/` altında saklanır. Aynı şablon yeniden üretildiğinde dosyaları yenilenir. Tarayıcı sekmesini kapatmak sunucuyu veya üretimi durdurmaz; uygulamayı sol alttaki düğmeden kapatın veya terminalde Ctrl+C kullanın.

## Kendi sesinizi kullanma

**Ses** sekmesinde mevcut ses dosyasını yükleyin ya da tarayıcı mikrofon izni vererek 15 saniyelik kayıt alın. Kaydı erken bitirmek de mümkündür. FFmpeg, sesi mono 24 kHz WAV biçimine dönüştürür. En fazla 50 MB kabul edilir. Yeni referans `sesim.wav` adıyla kaydedilir; önceki referans yalnızca dönüşüm başarılı olunca değiştirilir. Kayıt Git’e dahil edilmez.

Tarayıcı mikrofonu kullanıldığı için kayıt için Python `record` ek paketleri gerekmez. Klonlama için `.[clone]` ek kurulumu gerekir; [kurulum ayrıntıları](setup.md). **Klonlanmış sesi dene** kısa bir örnek üretir ve GUI’de dinlemeye açar. Ardından Üretim ekranında **Kendi sesim** seçin. Referans olmadan bu mod başlatılmaz. Otomatik mod referans varsa klonlama, yoksa profesyonel ses kullanır.

## Ortam kontrolü ve komut satırı

**Ortam** sekmesindeki kontrol, bağımlılıkları günlükte listeler. Bir işlem devam ederken başka işlem başlatılmaz. Sunucu yalnızca `127.0.0.1` üzerinde, oturuma özel yerel adreste çalışır. Arayüz dosyaları yereldir; CDN veya çevrimiçi font kullanmaz.

```bash
python baslat.py --help
python baslat.py gui --no-browser
python baslat.py gui --port 8765
python baslat.py doctor
python baslat.py generate --preset 1 --mode pro
python baslat.py generate --all --mode pro
npm run doctor
```

Komut satırı üretim yolu otomasyon için korunmuştur; etkileşimli seçimler GUI’de yapılır. `tubie` ve `tubie record` sırasıyla üretim ve ses ekranını açar.

[Türkçe stil kataloğu](catalog/README.tr.md) ve [İngilizce katalog](catalog/README.en.md), skill üzerinden yönetilen Director/Autopilot akışlarını açıklar. GUI mevcut dört Türkçe siber güvenlik şablonunu yönetir; katalogdaki diğer stiller skill akışında kullanılır.
