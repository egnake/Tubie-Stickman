# Kurulum ve sorun giderme

Ana kurulum komutları [README](../README.md#kurulum) içindedir. Uygulama, Go araç zinciri ve Node.js motoru ile doğrudan çalıştırılmak üzere tasarlanmıştır.

## Go ve Node

Go 1.22+ ve Node.js 22 LTS ortak başlangıç sürümüdür. Go araç zinciri sistem paket yöneticisinden veya resmî Go sitesinden temin edilir.

Node motorunun bağımlılıkları kökteki npm workspace üzerinden `npm ci` ile kurulur. Kökteki `package-lock.json` tüm depo için kullanılır; skill altındaki kilit dosyası, skill tek başına dağıtıldığında kullanılır.

`go run . doctor` komutunun çıkış kodu gerekli bileşen eksikse `1`, kontroller başarılıysa `0` olur. Bu kontrol `node`, `ffmpeg`, `ffprobe`, `@napi-rs/canvas` ve `sharp` bileşenlerini doğrular.

## Ses ve GPU

Temel profesyonel seslendirme yolu (`pro`) Edge TTS üzerinden çalışır ve harici GPU gerektirmez. Kendi sesinizle üretim için GUI üzerinden mikrofon kaydı alabilir veya 15 saniyelik referans ses dosyası yükleyebilirsiniz.

## FFmpeg, ses ve fontlar

`ffmpeg` ve `ffprobe` PATH üzerinde olmalıdır. Gerekirse `FFMPEG` ve `FFPROBE` ortam değişkenlerine yürütülebilir dosyanın tam yolunu verin. `NODE` değişkeni Node yürütülebilir yolunu değiştirebilir. Yolun etrafına değişken değerinin bir parçası olarak tırnak eklemeyin.

GUI, tarayıcının mikrofonunu kullanır. Tarayıcı mikrofonu desteklemiyorsa Ses sekmesinden dosya yükleyin. Windows'ta masaüstü ve tarayıcı mikrofon izinlerini kontrol edin. Dosyadan ses yükleme seçeneği mikrofon gerektirmez.

A6 son işleme ve Node render motoru depodaki fontları kullanır; Windows veya macOS font klasörü gerekmez.

## Çevrimdışı kullanım

Profesyonel seslendirme (`pro`) Edge TTS hizmetine bağlanır. Node render testleri sentetik yerel medya kullanır. Gelişmiş skill stillerinin bazıları ayrıca yapılandırılan harici görsel/ses sağlayıcılarını kullanır; ayrıntılar stil kataloğundadır.
