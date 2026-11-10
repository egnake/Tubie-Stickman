#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Stickman Video Director - Siber Güvenlik & Yazılım Viral İçerik Üretim Bandı
- Tema: Cyber Dark Mode (Gece Mavisi #0A0E17 + Neon Matris Izgarası + Parlayan Çöp Adamlar)
- Çöp Adamlar: Çoklu Etkileşimli (Hacker vs Kurban vs Güvenlik Uzmanı)
- Siber Donanımlar: Laptop, Terminal, Kilitler, Güvenlik Kalkanı, Zararlı Yazılım, SQL Veritabanı
- Seslendirme:
  1) Kendi Klonlanmış Sesiniz (XTTS-v2 RTX 3060 GPU - Tekrar Korumalı & Hızlı Tempolu)
  2) Profesyonel Belgesel Spikeri (AhmetNeural - Yüksek Enerji & Kusursuz Diksiyon)
- Altyazılar: Alex Hormozi & MrBeast Stili (1-3 kelime, kelime kelime yanan Neon Yeşil karaoke efekti)
- Ses Tasarımı: Kesintisiz Cyber Synthwave Beat + Prosedürel Siber Efektler (Klavye, Alarm, Glitch, Kilit)

"""

import sys

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
        sys.stderr.reconfigure(encoding="utf-8")
    except Exception:
        pass

import re
import json
import shutil
import asyncio
import argparse
import subprocess
import tempfile
from pathlib import Path
from typing import List, Dict, Any, Tuple

from .runtime import ROOT_DIR, get_binary_path, cuda_device

# ==============================================================================
# 1. DİZİNLER VE YOLLAR
# ==============================================================================

BASE_DIR = ROOT_DIR
SKILL_DIR = BASE_DIR / "skills" / "directing-stickman-videos"
OUTPUT_DIR = BASE_DIR / "output"
PROJECTS_DIR = BASE_DIR / "projects"
SESIM_WAV = BASE_DIR / "sesim.wav"
BGM_CYBER = BASE_DIR / "assets" / "audio" / "bgm_cyber.wav"
BGM_LOFI = BASE_DIR / "assets" / "audio" / "bgm_lofi.wav"

FFMPEG_EXE = get_binary_path("ffmpeg")
FFPROBE_EXE = get_binary_path("ffprobe")
NODE_EXE = get_binary_path("node")


# ==============================================================================
# 2. VİRAL SİBER GÜVENLİK & YAZILIM SENARYOLARI
# ==============================================================================

CYBER_VIRAL_SCRIPTS = {
    "1": {
        "title": "Wi-Fi Şifreniz 10 Saniyede Nasıl Kırılır?",
        "slug": "wifi_sifresi_nasil_kirilir",
        "narration": (
            "Şu an bağlandığın kafe Wi-Fi ağı seni izliyor olabilir! "
            "Hackerlar şifreni tahmin etmeye çalışmaz. "
            "Deauth saldırısıyla seni bir saniyede ağdan düşürürler. "
            "Telefonun tekrar bağlanırken şifreli paket havada yakalanır! "
            "Buna dört yollu el sıkışma denir. "
            "Hacker bu paketi güçlü bir bilgisayarda saniyeler içinde çözer. "
            "Çözüm çok basit: Halka açık ağlarda asla şifresiz gezme ve mutlaka VPN kullan! "
            "Siber dünyada av değil, avcı ol. Takipte kal!"
        ),
        "scenes": [
            # Sahne 1: Kurban kafede telefona bakıyor, yanında hacker laptopla pusuda
            {
                "title": "KAFE WI-FI AĞI GÜVENLİ Mİ?",
                "color": "#00F0FF",
                "underline": True,
                "sfx": "keyboard",
                "stickmen": [
                    {"pose": 0, "x": 130, "y": 790, "w": 470, "action": "bounce"},
                    {"pose": 1, "x": 680, "y": 800, "w": 470, "action": "shake"}
                ],
                "prop": {"path": "@b1/cyber-wifi.png", "x": 490, "y": 1420, "w": 300},
                "arrow": False
            },
            # Sahne 2: Saldırı anı: Ağdan koparma (Deauth)
            {
                "title": "BİR SANİYEDE AĞDAN DÜŞTÜN!",
                "color": "#FF3B30",
                "underline": False,
                "sfx": "alert",
                "stickmen": [
                    {"pose": 2, "x": 130, "y": 810, "w": 480, "action": "shake"},
                    {"pose": 1, "x": 670, "y": 780, "w": 480, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-skull.png", "x": 490, "y": 1420, "w": 300},
                "arrow": True
            },
            # Sahne 3: Havada paket yakalama (Handshake capture)
            {
                "title": "ŞİFRELİ PAKET YAKALANDI!",
                "color": "#FFE500",
                "underline": True,
                "sfx": "glitch",
                "stickmen": [
                    {"pose": 1, "x": 140, "y": 790, "w": 480, "action": "bounce"},
                    {"pose": 2, "x": 670, "y": 810, "w": 460, "action": "shake"}
                ],
                "prop": {"path": "@b1/cyber-laptop.png", "x": 480, "y": 1420, "w": 310},
                "arrow": False
            },
            # Sahne 4: Şifre Çözüldü
            {
                "title": "ŞİFRE ÇÖZÜLDÜ: ERİŞİM AÇIK!",
                "color": "#FF3B30",
                "underline": False,
                "sfx": "alert",
                "stickmen": [
                    {"pose": 2, "x": 120, "y": 810, "w": 480, "action": "shake"},
                    {"pose": 0, "x": 660, "y": 790, "w": 480, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-padlock-unlocked.png", "x": 490, "y": 1420, "w": 290},
                "arrow": True
            },
            # Sahne 5: Çözüm: VPN ve Güvenli Protokol
            {
                "title": "KESİN ÇÖZÜM: VPN & WPA3",
                "color": "#00FF66",
                "underline": True,
                "sfx": "chime",
                "stickmen": [
                    {"pose": 1, "x": 150, "y": 780, "w": 490, "action": "bounce"},
                    {"pose": 0, "x": 670, "y": 800, "w": 460, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-shield.png", "x": 490, "y": 1420, "w": 300},
                "arrow": False
            },
            # Sahne 6: Call to Action: Takipte kal
            {
                "title": "GÜVENDE KALMAK İÇİN TAKİP ET!",
                "color": "#00FF66",
                "underline": True,
                "sfx": "jingle",
                "stickmen": [
                    {"pose": 1, "x": 140, "y": 770, "w": 500, "action": "bounce"},
                    {"pose": 2, "x": 650, "y": 770, "w": 500, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-terminal.png", "x": 480, "y": 1420, "w": 310},
                "arrow": False
            }
        ]
    },
    "2": {
        "title": "Hackerlar Şifreni 3 Saniyede Nasıl Kırıyor?",
        "slug": "sifren_nasil_kirilir_brute_force",
        "narration": (
            "Eğer şifren bir iki üç dört beş altı veya doğum tarihinse, hackerların sana ihtiyacı bile yok! "
            "Sıradan bir hacker bilgisayarı saniyede milyarlarca parola dener. "
            "Buna Kaba Kuvvet yani Brute Force saldırısı denir. "
            "Sekiz karakterli basit bir şifre güçlü bir ekran kartıyla sadece üç saniyede kırılır! "
            "Büyük harf, küçük harf, rakam ve sembol kullanmak bu süreyi yüzlerce yıla çıkarır. "
            "Parolanı hemen güncelle, güvenliğini şansa bırakma!"
        ),
        "scenes": [
            {
                "title": "ŞİFREN GÜVENDE Mİ?",
                "color": "#FF3B30",
                "underline": True,
                "sfx": "keyboard",
                "stickmen": [
                    {"pose": 0, "x": 130, "y": 790, "w": 470, "action": "bounce"},
                    {"pose": 2, "x": 670, "y": 810, "w": 470, "action": "shake"}
                ],
                "prop": {"path": "@b1/cyber-padlock-unlocked.png", "x": 490, "y": 1420, "w": 290},
                "arrow": False
            },
            {
                "title": "SANİYEDE MİLYARLARCA DENEME!",
                "color": "#FFE500",
                "underline": False,
                "sfx": "alert",
                "stickmen": [
                    {"pose": 1, "x": 130, "y": 780, "w": 480, "action": "bounce"},
                    {"pose": 2, "x": 670, "y": 810, "w": 470, "action": "shake"}
                ],
                "prop": {"path": "@b1/cyber-laptop.png", "x": 490, "y": 1420, "w": 300},
                "arrow": True
            },
            {
                "title": "BRUTE FORCE SALDIRISI!",
                "color": "#FF3B30",
                "underline": True,
                "sfx": "glitch",
                "stickmen": [
                    {"pose": 2, "x": 110, "y": 810, "w": 470, "action": "shake"},
                    {"pose": 1, "x": 670, "y": 780, "w": 480, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-skull.png", "x": 480, "y": 1420, "w": 300},
                "arrow": False
            },
            {
                "title": "SADECE 3 SANİYE SÜRDÜ!",
                "color": "#FF3B30",
                "underline": False,
                "sfx": "alert",
                "stickmen": [
                    {"pose": 2, "x": 120, "y": 820, "w": 480, "action": "shake"},
                    {"pose": 0, "x": 660, "y": 790, "w": 470, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-database.png", "x": 490, "y": 1420, "w": 300},
                "arrow": True
            },
            {
                "title": "GÜÇLÜ PAROLA = YÜZLERCE YIL",
                "color": "#00FF66",
                "underline": True,
                "sfx": "chime",
                "stickmen": [
                    {"pose": 1, "x": 140, "y": 780, "w": 490, "action": "bounce"},
                    {"pose": 0, "x": 670, "y": 800, "w": 460, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-padlock-locked.png", "x": 490, "y": 1420, "w": 290},
                "arrow": False
            },
            {
                "title": "ŞİFRENİ HEMEN GÜNCELLE!",
                "color": "#00FF66",
                "underline": True,
                "sfx": "lock",
                "stickmen": [
                    {"pose": 1, "x": 130, "y": 770, "w": 500, "action": "bounce"},
                    {"pose": 2, "x": 660, "y": 770, "w": 500, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-shield.png", "x": 480, "y": 1420, "w": 300},
                "arrow": False
            }
        ]
    },
    "3": {
        "title": "Dünyanın En Tehlikeli Yazılım Açığı: SQL Injection!",
        "slug": "sql_injection_nasil_calisir",
        "narration": (
            "Giriş kutusuna sadece on karakter yazarak milyonlarca kullanıcının verisi çalınabilir mi? "
            "Evet! Bunun adı SQL Injection! "
            "Kötü kodlanmış bir web sitesinde kullanıcı adı yerine tek tırnak veya bir eşittir bir yazarsan... "
            "Veritabanı sorgusu bozulur ve kilit doğrudan açılır! "
            "Tüm kullanıcı adları, şifreler ve kredi kartları ekrana akar. "
            "Yazılımcılar için çözüm tek kelime: Parametreli sorgu kullanmak! "
            "Güvenli kod yaz, sistemini koru!"
        ),
        "scenes": [
            {
                "title": "10 KARAKTERLE BÜYÜK VURGUN!",
                "color": "#FF3B30",
                "underline": True,
                "sfx": "keyboard",
                "stickmen": [
                    {"pose": 1, "x": 130, "y": 790, "w": 480, "action": "bounce"},
                    {"pose": 2, "x": 670, "y": 810, "w": 470, "action": "shake"}
                ],
                "prop": {"path": "@b1/cyber-terminal.png", "x": 480, "y": 1420, "w": 310},
                "arrow": False
            },
            {
                "title": "SQL INJECTION TUZAĞI",
                "color": "#FFE500",
                "underline": False,
                "sfx": "glitch",
                "stickmen": [
                    {"pose": 1, "x": 140, "y": 780, "w": 480, "action": "bounce"},
                    {"pose": 0, "x": 670, "y": 800, "w": 460, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-database.png", "x": 490, "y": 1420, "w": 300},
                "arrow": True
            },
            {
                "title": "VERİTABANI KİLİDİ PATLADI!",
                "color": "#FF3B30",
                "underline": True,
                "sfx": "alert",
                "stickmen": [
                    {"pose": 2, "x": 110, "y": 820, "w": 480, "action": "shake"},
                    {"pose": 1, "x": 670, "y": 780, "w": 480, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-bug.png", "x": 480, "y": 1420, "w": 300},
                "arrow": False
            },
            {
                "title": "TÜM ŞİFRELER ELE GEÇİRİLDİ!",
                "color": "#FF3B30",
                "underline": False,
                "sfx": "alert",
                "stickmen": [
                    {"pose": 2, "x": 120, "y": 810, "w": 480, "action": "shake"},
                    {"pose": 0, "x": 660, "y": 790, "w": 480, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-skull.png", "x": 490, "y": 1420, "w": 300},
                "arrow": True
            },
            {
                "title": "ÇÖZÜM: PARAMETRELİ SORGU",
                "color": "#00FF66",
                "underline": True,
                "sfx": "chime",
                "stickmen": [
                    {"pose": 1, "x": 140, "y": 780, "w": 490, "action": "bounce"},
                    {"pose": 0, "x": 670, "y": 800, "w": 460, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-shield.png", "x": 490, "y": 1420, "w": 300},
                "arrow": False
            },
            {
                "title": "GÜVENLİ KOD YAZ, TAKİPTE KAL!",
                "color": "#00FF66",
                "underline": True,
                "sfx": "jingle",
                "stickmen": [
                    {"pose": 1, "x": 130, "y": 770, "w": 500, "action": "bounce"},
                    {"pose": 2, "x": 660, "y": 770, "w": 500, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-padlock-locked.png", "x": 480, "y": 1420, "w": 290},
                "arrow": False
            }
        ]
    },
    "4": {
        "title": "Tek Bir Sahte Linke Tıkladın ve Her Şey Gitti! (Oltalama / Phishing)",
        "slug": "phishing_oltalama_tuzagi",
        "narration": (
            "Bankandan gelen acil bir mesaj veya kargo bildirimi... "
            "Tek bir tıkla tüm hesapların dakikalar içinde boşaltılabilir! "
            "Oltalama saldırılarında hackerlar gerçeğinin birebir kopyası sahte siteler kurar. "
            "Kullanıcı adı ve şifreni girdiğin anda doğrudan hackerın ekranına düşer! "
            "Giriş yapmadan önce web adresini harf harf kontrol et. "
            "Şüpheli hiçbir linke tıklama, güvenliğini koru!"
        ),
        "scenes": [
            {
                "title": "TEK BİR TIKLA HER ŞEY BİTTİ!",
                "color": "#FF3B30",
                "underline": True,
                "sfx": "pop",
                "stickmen": [
                    {"pose": 0, "x": 130, "y": 790, "w": 470, "action": "bounce"},
                    {"pose": 2, "x": 670, "y": 810, "w": 470, "action": "shake"}
                ],
                "prop": {"path": "@b1/cyber-wifi.png", "x": 490, "y": 1420, "w": 300},
                "arrow": False
            },
            {
                "title": "SAHTE SİTE TUZAĞI (PHISHING)",
                "color": "#FFE500",
                "underline": False,
                "sfx": "keyboard",
                "stickmen": [
                    {"pose": 1, "x": 130, "y": 780, "w": 480, "action": "bounce"},
                    {"pose": 0, "x": 670, "y": 800, "w": 460, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-laptop.png", "x": 480, "y": 1420, "w": 310},
                "arrow": True
            },
            {
                "title": "ŞİFREN SALDIRGANDA!",
                "color": "#FF3B30",
                "underline": True,
                "sfx": "glitch",
                "stickmen": [
                    {"pose": 2, "x": 110, "y": 820, "w": 480, "action": "shake"},
                    {"pose": 1, "x": 670, "y": 780, "w": 480, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-terminal.png", "x": 490, "y": 1420, "w": 300},
                "arrow": False
            },
            {
                "title": "HESAPLAR BOŞALTILDI!",
                "color": "#FF3B30",
                "underline": False,
                "sfx": "alert",
                "stickmen": [
                    {"pose": 2, "x": 120, "y": 810, "w": 480, "action": "shake"},
                    {"pose": 0, "x": 660, "y": 790, "w": 480, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-skull.png", "x": 490, "y": 1420, "w": 300},
                "arrow": True
            },
            {
                "title": "URL ADRESİNİ MUTLAKA KONTROL ET!",
                "color": "#00FF66",
                "underline": True,
                "sfx": "chime",
                "stickmen": [
                    {"pose": 1, "x": 140, "y": 780, "w": 490, "action": "bounce"},
                    {"pose": 0, "x": 670, "y": 800, "w": 460, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-shield.png", "x": 490, "y": 1420, "w": 300},
                "arrow": False
            },
            {
                "title": "SİBER FARKINDALIK HAYAT KURTARIR!",
                "color": "#00FF66",
                "underline": True,
                "sfx": "jingle",
                "stickmen": [
                    {"pose": 1, "x": 130, "y": 770, "w": 500, "action": "bounce"},
                    {"pose": 2, "x": 660, "y": 770, "w": 500, "action": "bounce"}
                ],
                "prop": {"path": "@b1/cyber-padlock-locked.png", "x": 480, "y": 1420, "w": 290},
                "arrow": False
            }
        ]
    }
}


# ==============================================================================
# 3. METİN NORMALİZASYONU (SIFIR KELİME TEKRARI GARANTİSİ)
# ==============================================================================

def normalize_turkish_text(text: str) -> str:
    """
    Sayıları, sembolleri ve noktalama işaretlerini TTS için temizler.
    XTTS'in autoregressive decoder döngüsüne ('gün gün') girmesini engeller.
    """
    replacements = [
        ("%100", "yüzde yüz"), ("%50", "yüzde elli"), ("%1", "yüzde bir"),
        ("%", "yüzde "),
        ("10", "on"), ("9", "dokuz"), ("8", "sekiz"), ("7", "yedi"),
        ("6", "altı"), ("5", "beş"), ("4", "dört"), ("3", "üç"),
        ("2", "iki"), ("1", "bir"), ("0", "sıfır"),
        ("WPA3", "ve pe a üç"), ("WPA2", "ve pe a iki"),
        ("VPN", "ve pe ne"), ("2FA", "iki faktörlü doğrulama"),
        ("SQL", "es kyu el"), ("URL", "u re le"), ("SMS", "se me se"),
        ("WiFi", "vayfay"), ("Wi-Fi", "vayfay")
    ]
    cleaned = text
    for old, new in replacements:
        cleaned = re.sub(rf"\b{re.escape(old)}\b", new, cleaned, flags=re.IGNORECASE)

    # Fazla boşlukları ve gereksiz karakterleri temizle
    cleaned = re.sub(r"\s+", " ", cleaned).strip()
    return cleaned


# ==============================================================================
# 4. SESLENDİRME MOTORU (KLONLANMIŞ SES VEYA PROFESYONEL BELGESEL SPİKERİ)
# ==============================================================================

class VoiceEngine:
    def __init__(self, mode: str = "auto", voice_name: str = "tr-TR-AhmetNeural"):
        """
        mode:
          - 'cloned': Kendi sesinizi XTTS-v2 ile klonlar
          - 'pro': Edge-TTS AhmetNeural ile yüksek enerjili siber belgesel spikeri
          - 'auto': sesim.wav varsa klonlar, yoksa pro modunu kullanır
        """
        self.mode = mode
        self.voice_name = voice_name
        self._xtts_model = None

    def _get_xtts(self):
        if self._xtts_model is None:
            try:
                from TTS.api import TTS
            except ImportError as exc:
                raise RuntimeError("XTTS için coqui-tts kurulmalı (voice bağımlılık grubu).") from exc
            device = cuda_device()
            print(f"⚡ Coqui XTTS-v2 yükleniyor (Aygıt: {device.upper()})...")
            self._xtts_model = TTS("tts_models/multilingual/multi-dataset/xtts_v2").to(device)
        return self._xtts_model

    def synthesize(self, text: str, out_wav: Path) -> Tuple[float, str]:
        normalized = normalize_turkish_text(text)
        out_wav.parent.mkdir(parents=True, exist_ok=True)

        use_cloned = False
        if self.mode == "cloned":
            use_cloned = True
        elif self.mode == "auto" and SESIM_WAV.is_file():
            use_cloned = True

        if self.mode == "cloned" and not SESIM_WAV.is_file():
            raise FileNotFoundError(f"Klonlanmış ses için referans kaydı bulunamadı: {SESIM_WAV}")

        if use_cloned and SESIM_WAV.is_file():
            print(f"🚀 KENDİ SESİNİZ ('sesim.wav') KULLANILARAK XTTS-v2 İLE KLONLANIYOR...")
            print(f"🛡️  Tekrar önleme filtresi (repetition_penalty=4.5, speed=1.08) devrede...")
            tts = self._get_xtts()
            tts.tts_to_file(
                text=normalized,
                speaker_wav=str(SESIM_WAV),
                language="tr",
                file_path=str(out_wav),
                split_sentences=True,
                speed=1.08,
                repetition_penalty=4.5,
                temperature=0.55,
                top_p=0.8
            )
            mode_used = "cloned"
        else:
            print(f"🎙️ PROFESYONEL SİBER BELGESEL SPİKERİ ({self.voice_name}, rate=+12%)...")
            async def run_edge(mp3_path):
                try:
                    import edge_tts
                except ImportError as exc:
                    raise RuntimeError("Profesyonel ses için edge-tts kurulmalı.") from exc
                communicate = edge_tts.Communicate(
                    normalized,
                    self.voice_name,
                    rate="+12%",
                    pitch="+0Hz"
                )
                await communicate.save(str(mp3_path))
            with tempfile.TemporaryDirectory(prefix="tubie-tts-") as temp_dir:
                mp3_path = Path(temp_dir) / "narration.mp3"
                asyncio.run(run_edge(mp3_path))
                subprocess.run([FFMPEG_EXE, "-v", "error", "-y", "-i", str(mp3_path),
                                "-ar", "24000", "-ac", "1", str(out_wav)], check=True)
            mode_used = "pro"

        duration = self._get_duration(out_wav)
        print(f"✅ Seslendirme tamamlandı! ({mode_used.upper()}) Süre: {duration:.3f} saniye")
        return duration, mode_used

    def _get_duration(self, wav_path: Path) -> float:
        cmd = [
            FFPROBE_EXE, "-v", "error",
            "-show_entries", "format=duration",
            "-of", "default=noprint_wrappers=1:nokey=1",
            str(wav_path)
        ]
        res = subprocess.run(cmd, capture_output=True, text=True, check=True)
        return float(res.stdout.strip())


# ==============================================================================
# 5. KELİME ZAMANLAMALARI VE VİRAL ALTYAZI KURGUSU
# ==============================================================================

class WordAligner:
    """Faster-Whisper GPU hızlandırması ile her kelimenin ses zamanını çıkarır."""
    def __init__(self):
        self._model = None

    def _get_whisper(self):
        if self._model is None:
            try:
                from faster_whisper import WhisperModel
            except ImportError as exc:
                raise RuntimeError("Altyazı hizalama için faster-whisper kurulmalı.") from exc
            device = cuda_device()
            compute_type = "float16" if device == "cuda" else "int8"
            print(f"🧠 Faster-Whisper GPU kelime zamanlayıcı hazırlanıyor ({device})...")
            self._model = WhisperModel("base", device=device, compute_type=compute_type)
        return self._model

    def align(self, wav_path: Path) -> List[Dict[str, Any]]:
        model = self._get_whisper()
        segments, _ = model.transcribe(str(wav_path), language="tr", word_timestamps=True)
        words = []
        for segment in segments:
            for w in segment.words:
                cleaned = re.sub(r"[^\wçğıöşüÇĞİÖŞÜ]", "", w.word).strip()
                if cleaned:
                    words.append({
                        "word": cleaned,
                        "start": round(w.start, 3),
                        "end": round(w.end, 3)
                    })
        return words

    def build_viral_captions(self, words: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """
        Kelimeleri 1-3'erli gruplara böler ve her grubun içine kelime kelime zamanlamasını
        ('words' dizisi) ekler. Böylece render motoru aktif kelimeyi NEON renkle yakabilir!
        """
        cues = []
        i = 0
        n = len(words)
        while i < n:
            take = 2
            if i + 3 <= n:
                total_len = sum(len(words[k]["word"]) for k in range(i, i + 3))
                if total_len <= 18:
                    take = 3
                elif len(words[i]["word"]) > 9:
                    take = 1
            elif i + 1 == n:
                take = 1

            group = words[i:i + take]
            start_t = group[0]["start"]
            end_t = group[-1]["end"]

            # Zamanlama sırasını garantiye al
            if cues and start_t < cues[-1]["end"]:
                start_t = cues[-1]["end"]
            if end_t <= start_t:
                end_t = round(start_t + 0.28, 3)

            full_text = " ".join(w["word"] for w in group)
            cues.append({
                "start": start_t,
                "end": end_t,
                "text": full_text.upper(),
                "words": [
                    {
                        "word": w["word"].upper(),
                        "start": w["start"],
                        "end": w["end"]
                    }
                    for w in group
                ]
            })
            i += take
        return cues


# ==============================================================================
# 6. SİBER SAHNE & ÇOKLU ÇÖP ADAM YÖNETMENİ (STORYBOARD DIRECTOR)
# ==============================================================================

class CyberStoryboardDirector:
    """
    Siber Güvenlik estetiğine uygun sahne koreografisi:
    - Dark Cyber Theme (#0A0E17, yeşil matris ızgara)
    - Beyaz / parlayan çoklu çöp adamlar
    - Siber donanım propları (Laptop, Terminal, Kilit, Kalkan, Veritabanı)
    - Prosedürel SFX (Klavye, Alarm, Glitch, Kilit)
    """
    def __init__(self, scenes: List[Dict[str, Any]], total_duration: float):
        self.scenes = scenes
        self.total_duration = total_duration

    def build_storyboard(self) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]]]:
        num_scenes = len(self.scenes)
        scene_duration = self.total_duration / num_scenes

        shots = []
        sfx = []

        for idx, scene in enumerate(self.scenes):
            start = round(idx * scene_duration, 3)
            end = round(self.total_duration if idx == num_scenes - 1 else (idx + 1) * scene_duration, 3)
            shot_len = end - start

            layers = []

            # 1. Siber Başlık Metni (Üst Kısım, Yüksek Kontrast)
            title_text = scene.get("title", "").upper()
            title_color = scene.get("color", "#00F0FF")
            layers.append({
                "type": "text",
                "text": title_text,
                "x": 80,
                "y": 280,
                "w": 1120,
                "size": 58,
                "bold": True,
                "color": title_color,
                "keys": [
                    {"t": 0, "alpha": 0.0},
                    {"t": 0.2, "alpha": 1.0}
                ]
            })

            # 2. Vurgu Çizgisi
            if scene.get("underline", True):
                layers.append({
                    "type": "underline",
                    "x": 80,
                    "y": 360,
                    "w": min(1000, max(360, len(title_text) * 32)),
                    "color": title_color,
                    "lineWidth": 7,
                    "keys": [
                        {"t": 0.05, "progress": 0.0},
                        {"t": 0.45, "progress": 1.0}
                    ]
                })

            # 3. Çoklu Çöp Adamlar (Multi-Stickman: 2 veya 3 karakter)
            stickmen = scene.get("stickmen", [{"pose": 0, "x": 400, "y": 800, "w": 480, "action": "bounce"}])
            for s_idx, sm in enumerate(stickmen):
                pose = sm.get("pose", 0)
                base_x = sm.get("x", 160 + s_idx * 500)
                base_y = sm.get("y", 800)
                base_w = sm.get("w", 480)
                action = sm.get("action", "bounce")

                if action == "shake":
                    # Stres / hacklenme / panik sarsıntısı
                    keys = [
                        {"t": 0, "x": base_x, "y": base_y, "w": base_w},
                        {"t": 0.08, "x": base_x - 14, "y": base_y, "w": base_w},
                        {"t": 0.16, "x": base_x + 14, "y": base_y, "w": base_w},
                        {"t": 0.24, "x": base_x - 8, "y": base_y, "w": base_w},
                        {"t": 0.32, "x": base_x + 8, "y": base_y, "w": base_w},
                        {"t": 0.40, "x": base_x, "y": base_y, "w": base_w}
                    ]
                else:
                    # Ritmik konuşma ve yaylanma
                    half = min(1.0, shot_len / 2)
                    keys = [
                        {"t": 0, "x": base_x, "y": base_y, "w": base_w},
                        {"t": round(half * 0.5, 3), "x": base_x, "y": base_y - 30, "w": base_w + 12},
                        {"t": round(half, 3), "x": base_x, "y": base_y, "w": base_w}
                    ]

                layers.append({
                    "type": "pose",
                    "pose": pose,
                    "x": base_x,
                    "y": base_y,
                    "w": base_w,
                    "keys": keys
                })

            # 4. Siber Prop (Laptop, Kalkan, Kilit, Terminal, Veritabanı vb.)
            prop_info = scene.get("prop")
            if prop_info:
                p_path = prop_info["path"]
                p_x = prop_info.get("x", 480)
                p_y = prop_info.get("y", 1420)
                p_w = prop_info.get("w", 300)
                layers.append({
                    "type": "image",
                    "path": p_path,
                    "x": p_x,
                    "y": p_y,
                    "w": p_w,
                    "keys": [
                        {"t": 0, "w": 80, "alpha": 0.0, "angle": 0.0},
                        {"t": 0.25, "w": p_w + 35, "alpha": 1.0, "angle": 0.05},
                        {"t": 0.42, "w": p_w, "alpha": 1.0, "angle": 0.0}
                    ]
                })

            # 5. Siber Animasyonlu Çizilen Ok
            if scene.get("arrow", False):
                layers.append({
                    "type": "arrow",
                    "x": 380,
                    "y": 1490,
                    "w": 520,
                    "color": "#FF3B30",
                    "keys": [{"t": 0, "progress": 0}, {"t": 0.45, "progress": 1}]
                })

            shots.append({
                "start": start,
                "end": end,
                "layers": layers
            })

            # 6. Ritmik Siber SFX (tap, pop, chime, jingle, paper)
            raw_sfx = scene.get("sfx", "tap")
            sfx_map = {
                "keyboard": "tap",
                "alert": "chime",
                "glitch": "paper",
                "lock": "jingle",
                "pop": "pop",
                "chime": "chime",
                "jingle": "jingle",
                "tap": "tap",
                "paper": "paper"
            }
            sfx_kind = sfx_map.get(raw_sfx, "tap")
            sfx.append({
                "time": round(start + 0.08, 3),
                "kind": sfx_kind,
                "gain": 0.25
            })

        return shots, sfx


# ==============================================================================
# 7. VİRAL RENDER MOTORU (AUTOPILOT ENTEGRASYONU)
# ==============================================================================

class CyberVideoPipeline:
    def __init__(self, mode: str = "auto", voice_name: str = "tr-TR-AhmetNeural", reporter=None, review=None):
        self.voice_engine = VoiceEngine(mode=mode, voice_name=voice_name)
        self.aligner = WordAligner()
        self.reporter = reporter
        self.review = review

    def report(self, stage, label):
        if self.reporter:
            self.reporter(stage=stage, label=label)

    def produce(self, script_data: Dict[str, Any]) -> Path:
        title = script_data["title"]
        slug = script_data["slug"]
        narration = script_data["narration"]
        scenes = script_data["scenes"]

        print(f"\n{'='*65}")
        print(f"🎬 SİBER GÜVENLİK VİRAL VİDEO MOTORU: {title}")
        print(f"{'='*65}")

        OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
        project_dir = PROJECTS_DIR / slug
        project_dir.mkdir(parents=True, exist_ok=True)

        # 1. Seslendirme Üret
        self.report('voice', 'Seslendirme hazırlanıyor')
        narration_wav = project_dir / "narration.wav"
        duration, mode_used = self.voice_engine.synthesize(narration, narration_wav)

        # 2. Faster-Whisper ile Kelime Hizalama & Viral Altyazılar
        self.report('captions', 'Kelime zamanlamaları çıkarılıyor')
        print(f"🧠 Faster-Whisper GPU ile kelime zamanlamaları çıkarılıyor...")
        words = self.aligner.align(narration_wav)
        captions = self.aligner.build_viral_captions(words)
        print(f"✅ Toplam {len(captions)} adet ritmik karaoke altyazı bloğu hazırlandı.")

        # 3. Siber Sahne Kurgusu (Dark Cyber Theme)
        self.report('storyboard', 'Sahneler ve ses tasarımı hazırlanıyor')
        director = CyberStoryboardDirector(scenes, duration)
        shots, sfx = director.build_storyboard()

        # 4. Arka Plan Müziği (Cyber Synthwave)
        bgm_path = BGM_CYBER if BGM_CYBER.is_file() else BGM_LOFI
        bgm_relative = None
        if bgm_path.is_file():
            shutil.copy(bgm_path, project_dir / "bgm.wav")
            bgm_relative = {"path": "bgm.wav", "gain": 0.12}

        # 5. project.json Dosyasını Yapılandır
        project_config = {
            "version": 1,
            "style": "B1",
            "theme": "cyber",
            "backgroundColor": "#0A0E17",
            "width": 720,
            "height": 1280,
            "fps": 30,
            "duration": duration,
            "captionStyle": "viral",
            "captionSize": 56,
            "captionY": 1460,
            "captionColor": "#00FF66",
            "captionActiveColor": "#00FF66",
            "captionBg": "rgba(7, 12, 20, 0.96)",
            "captionBorder": "rgba(0, 255, 128, 0.50)",
            "voice": {"path": "narration.wav", "lead": 0.0},
            "shots": shots,
            "captions": captions,
            "sfx": sfx
        }
        if bgm_relative:
            project_config["music"] = bgm_relative

        with open(project_dir / "project.json", "w", encoding="utf-8") as f:
            json.dump(project_config, f, ensure_ascii=False, indent=2)

        # 6. Render Aşaması
        autopilot_script = SKILL_DIR / "scripts" / "autopilot.mjs"
        live = ['--live'] if self.reporter else []
        self.report('preview', 'Önizleme oluşturuluyor')
        print(f"[1/3] Video önizlemesi oluşturuluyor...")
        subprocess.run([NODE_EXE, str(autopilot_script), "preview", str(project_dir / "project.json"), "--replace", *live], check=True)

        if self.review:
            self.review(project_dir)
        print(f"[2/3] Önizleme onaylanıyor...")
        note = 'Kullanıcı GUI üzerinden önizlemeyi onayladı' if self.review else 'Otomatik Onay: Siber Guvenlik Viral Video'
        subprocess.run([NODE_EXE, str(autopilot_script), "approve", str(project_dir / "project.json"), "--note", note], check=True)

        self.report('render', 'Son video işleniyor')
        print(f"[3/3] Tam video MP4 olarak render ediliyor...")
        subprocess.run([NODE_EXE, str(autopilot_script), "render", str(project_dir / "project.json"), "--full", "--replace", *live], check=True)

        # 7. Final Dosyaları output Klasörüne Kopyala
        self.report('export', 'Video ve altyazılar kaydediliyor')
        film_mp4 = project_dir / "output" / "film.mp4"
        final_video_path = OUTPUT_DIR / f"{slug}.mp4"
        final_srt_path = OUTPUT_DIR / f"{slug}.srt"
        final_sheet_path = OUTPUT_DIR / f"{slug}-onizleme.png"

        shutil.copy(film_mp4, final_video_path)
        if (project_dir / "output" / "film.srt").is_file():
            shutil.copy(project_dir / "output" / "film.srt", final_srt_path)
        contact_sheet = project_dir / "output" / "film-contact-sheet.png"
        if contact_sheet.is_file():
            shutil.copy(contact_sheet, final_sheet_path)

        print(f"\n🎉 VİRAL SİBER VİDEO MÜKEMMEL ŞEKİLDE ÜRETİLDİ!")
        print(f"📁 Video:    {final_video_path}")
        print(f"📄 Altyazı:  {final_srt_path}")
        print(f"🖼️ Önizleme: {final_sheet_path}")
        print(f"🎤 Ses Modu: {'KENDİ KLONLANMIŞ SESİNİZ' if mode_used == 'cloned' else 'PROFESYONEL BELGESEL SPİKERİ'}")
        print(f"🎨 Tema:     DARK CYBER MODE (#0A0E17 + NEON YEŞİL KARAOKE)")
        print(f"⚡ Süre:     {duration:.2f} sn | Çözünürlük: 720x1280 (9:16 Dikey)")
        print(f"{'='*65}\n")

        return final_video_path


# ==============================================================================
# 8. CLI KOMUT SATIRI ARAYÜZÜ
# ==============================================================================

def main(argv=None):
    parser = argparse.ArgumentParser(description="Stickman Siber Güvenlik & Yazılım Viral Video Motoru")
    parser.add_argument("--preset", choices=["1", "2", "3", "4"], help="Hazır şablon numarası (1-4)")
    parser.add_argument("--mode", choices=["auto", "cloned", "pro"], default="auto", help="Ses modu: auto, cloned veya pro")
    parser.add_argument("--all", action="store_true", help="Tüm hazır siber şablonları sırayla render et")

    args = parser.parse_args(argv)
    pipeline = CyberVideoPipeline(mode=args.mode)

    if args.all:
        for k in ["1", "2", "3", "4"]:
            pipeline.produce(CYBER_VIRAL_SCRIPTS[k])
    elif args.preset:
        pipeline.produce(CYBER_VIRAL_SCRIPTS[args.preset])
    else:
        # Varsayılan olarak 1. Preset (Wi-Fi Şifresi)
        pipeline.produce(CYBER_VIRAL_SCRIPTS["1"])


if __name__ == "__main__":
    main()
