@echo off
chcp 65001 >nul
title Siber Güvenlik & Yazılım Viral Video Motoru (Shorts & Reels)

echo ===================================================================
echo 🛡️  SİBER GÜVENLİK & YAZILIM VİRAL İÇERİK ÜRETİM BANDI (Shorts ^& Reels)
echo 🎨 Tema: DARK CYBER MODE (#0A0E17 + Matrix Grid + Neon Karaoke)
echo ===================================================================
if exist sesim.wav (
    echo 🎤 Kendi Sesiniz: YÜKLÜ [sesim.wav]
) else (
    echo 🎙️  Kendi Sesiniz: Yok (Kendi sesiniz icin: sesimi_kaydet.bat)
)
echo ===================================================================
echo.
echo Lutfen uretmek istediginiz siber guvenlik videosunu secin:
echo   [1] Wi-Fi Sifreniz 10 Saniyede Nasil Kirilir? (Deauth ^& Handshake)
echo   [2] Hackerlar Sifreni 3 Saniyede Nasil Kiriyor? (Brute Force Tehlikesi)
echo   [3] Dunyanin En Tehlikeli Yazilim Acigi: SQL Injection!
echo   [4] Tek Bir Sahte Linke Tikladin ve Her Sey Gitti! (Phishing / Oltalama)
echo   [A] Tum siber videolari sirayla uret (Otomatik Seri Uretim)
echo   [K] Kendi Sesimi Yeniden Kaydet (Mikrofon)
echo.
set /p secim="Seciminiz (1/2/3/4/A/K) [Varsayilan 1]: "
if "%secim%"=="" set secim=1

if /i "%secim%"=="K" (
    python ses_klonlayici.py
    pause
    exit /b
)

echo.
echo Seslendirme Modunu Secin:
echo   [1] Profesyonel Siber Belgesel Spikeri (Tavsiye Edilen: Yuksek Enerji, Sifir Tekrar)
echo   [2] Kendi Klonlanmis Sesiniz (XTTS-v2 RTX 3060)
echo.
set /p ses_mod="Ses Modu (1 veya 2) [Varsayilan 1]: "

set MODE_ARG=--mode pro
if "%ses_mod%"=="2" (
    set MODE_ARG=--mode cloned
)

echo.
echo ===================================================================
echo 🚀 URETİM BASLATIYOR... LUTFEN BEKLEYIN...
echo ===================================================================

if /i "%secim%"=="A" (
    python turkce_video_uretici.py --all %MODE_ARG%
) else (
    python turkce_video_uretici.py --preset %secim% %MODE_ARG%
)

echo.
echo ===================================================================
echo 🎉 Ciktilar 'output' klasorune kaydedildi!
echo Klasor aciliyor...
explorer output
pause
