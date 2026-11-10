<div align="center">

# 🛡️ Tubie-Stickman

**Yapay Zeka Destekli Siber Güvenlik & Yazılım Viral İçerik Üretim Motoru**
*(YouTube Shorts · Instagram Reels · TikTok · Faceless Channels)*

[![GitHub Stars](https://img.shields.io/github/stars/egnake/Tubie-Stickman?style=flat-square)](https://github.com/egnake/Tubie-Stickman/stargazers)
[![License: MIT](https://img.shields.io/badge/License-MIT-16a34a?style=flat-square)](LICENSE)
[![Platform: 100% Local](https://img.shields.io/badge/Platform-100%25%20Local%20(0%20TL)-blue?style=flat-square)]()
[![Hardware: NVIDIA CUDA](https://img.shields.io/badge/GPU-NVIDIA%20RTX%20Accelerated-76b900?style=flat-square)]()

</div>

---

## ⚡ Nedir?

**Tubie-Stickman**, hiçbir bulut API ücreti veya aylık abonelik olmadan, **%100 yerel olarak bilgisayarınızda (GPU hızlandırmalı)** siber güvenlik ve yazılım konseptli viral shorts/reels videoları üreten tam teşekküllü bir içerik üretim bandıdır.

Gelişmiş çöp adam koreografisi, dinamik siber donanımlar (laptop, terminal, veritabanı, kilitler), MrBeast/Alex Hormozi tarzı kelime kelime yanan karaoke altyazılar ve kesintisiz synthwave arka plan müziği ile yayına hazır 9:16 dikey videolar oluşturur.

---

## 🚀 Öne Çıkan Özellikler

- **🎤 Kendi Sesinizi Klonlama (XTTS-v2 GPU):**
  - Mikrofonunuzdan sadece 10 saniyelik referans ses kaydı (`sesim.wav`) alarak tüm videoları kendi sesinizle seslendirir.
  - `repetition_penalty=4.5`, `speed=1.08` ve metin normalizasyonu sayesinde takılma veya kelime tekrarı yapmaz.
- **🎙️ Profesyonel Belgesel Spikeri (AhmetNeural + %12 Hız):**
  - Netflix/Vox teknoloji belgeselleri ayarında, yüksek tempolu ve kristal netliğinde alternatif seslendirme.
- **🎨 Dark Cyber Mode (#0A0E17):**
  - Karanlık gece mavisi arka plan, siber matris ızgarası ve karanlıkta parlayan beyaz çöp adamlar (`character-pose-sheet-cyber.png`).
- **💻 9 Farklı Siber Güvenlik Vektör Propu:**
  - Hacker Laptopu, Linux Terminali, AES-256 Kilitler, Güvenlik Kalkanı, Zararlı Yazılım (Bug), SQL Veritabanı ve Rogue Wi-Fi sinyali.
- **🔥 Alex Hormozi & MrBeast Karaoke Altyazıları:**
  - Faster-Whisper GPU ile milisaniyelik kelime zamanlaması.
  - O an okunan kelime parlayan **Neon Yeşil (`#00FF66`)** renkle yanar, diğer kelimeler kalın siyah konturlu beyaz kalır.
- **🎧 Cyber Synthwave Beat & Foley Ses Efektleri:**
  - 116 BPM kesintisiz bas ritimleri (`bgm_cyber.wav`).
  - Mekanik klavye tıkırtısı, alarm ve kilit ses efektleri.

---

## 📂 Hazır Siber Güvenlik Konseptleri

Sistemde tek tıkla üretilebilen hazır viral senaryolar:

1. **Wi-Fi Şifreniz 10 Saniyede Nasıl Kırılır?** *(Deauth & 4-Way Handshake Saldırısı)*
2. **Hackerlar Şifreni 3 Saniyede Nasıl Kırıyor?** *(Brute Force & GPU Hashcat)*
3. **Dünyanın En Tehlikeli Yazılım Açığı: SQL Injection!** *(Web Güvenliği & Parametreli Sorgular)*
4. **Tek Bir Sahte Linke Tıkladın ve Her Şey Gitti!** *(Phishing / Sosyal Mühendislik Tuzağı)*

---

## 🛠️ Kurulum ve Gereksinimler

- **İşletim Sistemi:** Windows 10/11 veya Linux
- **Ekran Kartı:** NVIDIA GPU (RTX 3060 veya üzeri önerilir)
- **Gereksinimler:** Node.js v20+, Python 3.10+, FFmpeg

```bash
# Depoyu klonlayın
git clone https://github.com/egnake/Tubie-Stickman.git
cd Tubie-Stickman

# Bağımlılıkları yükleyin
npm install
pip install torch torchaudio faster-whisper TTS edge-tts soundfile sounddevice
```

---

## 🎬 Tek Tıkla Kullanım

Windows üzerinde tek yapmanız gereken:

1. `uret.bat` dosyasına çift tıklayın.
2. Üretmek istediğiniz video numarasını seçin (`1`, `2`, `3`, `4` veya `A` ile hepsini seri üretin).
3. Seslendirme tercihini yapın (`1: Profesyonel Spiker`, `2: Kendi Klonlanmış Sesiniz`).
4. Render tamamlandığında `output` klasörü otomatik olarak açılacaktır!

---

## 📄 Lisans

Bu proje [MIT Lisansı](LICENSE) ile lisanslanmıştır.  
Geliştirici: **[egnake](https://github.com/egnake)**
