#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Stickman Video Director - Kendi Sesini Klonlama Aracı (Voice Cloner)
%100 Yerel, RTX 3060 GPU Destekli Coqui XTTS-v2 Entegrasyonu.
"""

import sys
import time
import subprocess
import tempfile
from pathlib import Path

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
        sys.stderr.reconfigure(encoding="utf-8")
    except Exception:
        pass

from .runtime import ROOT_DIR, get_binary_path, cuda_device

SESIM_WAV = ROOT_DIR / "sesim.wav"
FFMPEG_EXE = get_binary_path("ffmpeg")

# ==============================================================================
# OKUMA METNİ (Kullanıcının okuyacağı referans metin)
# ==============================================================================
REFERANS_METIN = (
    "Merhaba! Ben bu kaydı, yapay zeka videolarımda kendi ses tonumu klonlamak için yapıyorum. "
    "Bundan sonra üreteceğim tüm YouTube Shorts ve Instagram Reels videoları, "
    "tamamen benim sesimle, akıcı, enerjik ve profesyonel bir şekilde hazırlanacak. "
    "Hedeflerime ulaşmak ve kaliteli içerikler üretmek için şimdi harekete geçiyorum!"
)


def print_banner():
    print("\n" + "=" * 68)
    print("🎙️  STICKMAN VIDEO DIRECTOR - SES KLONLAMA MERKEZİ")
    print("   Yerel XTTS-v2: CUDA veya CPU")
    print("=" * 68)


def _convert_reference(source: Path):
    """Convert beside the destination and replace only after successful encoding."""
    SESIM_WAV.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix=".tubie-voice-", dir=SESIM_WAV.parent) as temp_dir:
        converted = Path(temp_dir) / "reference.wav"
        subprocess.run([
            FFMPEG_EXE, "-v", "error", "-y", "-i", str(source),
            "-ar", "24000", "-ac", "1", str(converted)
        ], check=True)
        converted.replace(SESIM_WAV)


def mikrofonla_kaydet(saniye: int = 15):
    """sounddevice kullanarak mikrofondan temiz ses kaydı alır."""
    try:
        import sounddevice as sd
        import soundfile as sf
        import numpy as np
    except ImportError:
        print("❌ HATA: 'sounddevice' veya 'soundfile' kütüphanesi eksik.")
        return False

    print("\n" + "-" * 68)
    print("📖 LÜTFEN AŞAĞIDAKİ METNİ DOĞAL VE NET BİR ŞEKİLDE SESLİ OKUYUN:")
    print("-" * 68)
    print(f"\n\"{REFERANS_METIN}\"\n")
    print("-" * 68)

    input("👉 Hazır olduğunuzda klavyeden [ENTER] tuşuna basın...")

    print("\nHazırlanın...")
    for i in range(3, 0, -1):
        print(f"  🔴 {i}...")
        time.sleep(1)

    samplerate = 24000
    print(f"\n🎙️  KAYIT BAŞLADI! ({saniye} saniye boyunca okuyun)...")

    recording = sd.rec(int(saniye * samplerate), samplerate=samplerate, channels=1, dtype='float32')
    
    for left in range(saniye, 0, -1):
        print(f"  ⏱️  Kalan süre: {left} sn", end="\r")
        time.sleep(1)
    
    sd.wait()
    print("\n\n✅ Kayıt tamamlandı! Ses işleniyor...")

    # Normalize et
    max_val = np.max(np.abs(recording))
    if max_val > 0.01:
        recording = recording / max_val * 0.90
    else:
        print("⚠️ Uyarı: Mikrofon sesi çok kısık veya sessiz algılandı.")

    with tempfile.TemporaryDirectory(prefix="tubie-recording-") as temp_dir:
        temp_rec = Path(temp_dir) / "recording.wav"
        sf.write(str(temp_rec), recording, samplerate)
        _convert_reference(temp_rec)

    print(f"🎉 Sesiniz başarıyla kaydedildi: {SESIM_WAV}")
    return True


def dosyadan_yukle():
    """Var olan bir ses dosyasını (WAV, MP3, M4A) dönüştürüp sesim.wav yapar."""
    print("\nBilgisayarınızdaki veya telefonunuzdaki ses kaydının tam yolunu girin:")
    print("Örnek: ~/Downloads/kayit.m4a veya C:\\Users\\Adiniz\\Downloads\\kayit.m4a")
    dosya_yolu = input("\nDosya yolu: ").strip().strip('"').strip("'")

    path = Path(dosya_yolu).expanduser()
    if not path.is_file():
        print(f"❌ Dosya bulunamadı: {dosya_yolu}")
        return False

    print("🔄 Ses dönüştürülüyor ve normalize ediliyor...")
    _convert_reference(path)

    print(f"🎉 Sesiniz başarıyla ayarlandı: {SESIM_WAV}")
    return True


def ses_testi_yap():
    """XTTS-v2 modeli ile klonlanan sesin 1 cümlelik testini üretir."""
    if not SESIM_WAV.is_file():
        print("❌ Önce ses kaydı oluşturmalısınız.")
        return False

    device = cuda_device()
    print(f"\n🧠 XTTS-v2 modeli yükleniyor ({device})...")
    try:
        from TTS.api import TTS
        tts = TTS("tts_models/multilingual/multi-dataset/xtts_v2").to(device)

        test_text = "Merhaba! Bu video tamamen benim klonlanmış sesimle üretildi. Nasıl çıkmış?"
        test_out = ROOT_DIR / "output" / "klonlanmis_ses_testi.wav"
        test_out.parent.mkdir(parents=True, exist_ok=True)

        print("🎙️ Klonlanmış sesinizle örnek test cümlesi üretiliyor...")
        tts.tts_to_file(
            text=test_text,
            speaker_wav=str(SESIM_WAV),
            language="tr",
            file_path=str(test_out)
        )
        print(f"\n✨ Klonlama testi harika bir şekilde tamamlandı!")
        print(f"🎧 Dinlemek için: {test_out}")
        print("Bundan sonra video üreticisini çalıştırdığınızda")
        print("otomatik olarak BU SESİNİZLE videolar üretilecektir!")
        return True
    except Exception as e:
        print(f"⚠️ Test sırasında bir hata oluştu: {e}")
        return False


def main(argv=None):
    import argparse
    argparse.ArgumentParser(description="Ses kaydı ve klonlama aracı").parse_args(argv)
    print_banner()

    if SESIM_WAV.is_file():
        print(f"ℹ️  Şu anda kayıtlı bir referans sesiniz mevcut ({SESIM_WAV.stat().st_size // 1024} KB).")
        print("   Yeni bir kayıt yaparak eskisini güncelleyebilirsiniz.\n")

    print("Lütfen bir seçenek seçin:")
    print("  [1] Mikrofondan Şimdi Canlı Kaydet (15 saniye okuma)")
    print("  [2] Var olan bir ses dosyasını kullan (WhatsApp, telefon, OBS kaydı vb.)")
    print("  [3] Klonlanmış sesle örnek test cümlesi üret ve dinle")
    print("  [4] Çıkış")

    secim = input("\nSeçiminiz (1/2/3/4) [Varsayılan 1]: ").strip() or "1"

    if secim in ("1", "2"):
        success = mikrofonla_kaydet(saniye=15) if secim == "1" else dosyadan_yukle()
        if not success:
            return 1
        cevap = input("\nŞimdi klonlanmış sesinizle hemen bir test cümlesi üretelim mi? (E/H) [E]: ").strip().upper()
        return 0 if cevap == "H" or ses_testi_yap() else 1
    if secim == "3":
        return 0 if ses_testi_yap() else 1
    if secim == "4":
        print("Çıkılıyor...")
        return 0
    print("Geçersiz seçim.")
    return 1


if __name__ == "__main__":
    sys.exit(main())
