#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Generate High-Definition Vector-style Cyber Security Props
"""
from pathlib import Path
from PIL import Image, ImageDraw

BASE_DIR = Path(__file__).resolve().parents[1]
ASSET_DIR = BASE_DIR / "skills" / "directing-stickman-videos" / "assets" / "b1"
ASSET_DIR.mkdir(parents=True, exist_ok=True)

# 1. Cyber Laptop
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.rounded_rectangle([60, 80, 452, 340], radius=16, fill="#111827", outline="#374151", width=6)
d.rounded_rectangle([76, 96, 436, 324], radius=8, fill="#030712", outline="#00FF66", width=3)
d.text((95, 115), "> root@kali:~# ./exploit.sh", fill="#00FF66")
d.text((95, 145), "[*] Scanning target port 443...", fill="#00F0FF")
d.text((95, 175), "[+] VULNERABILITY FOUND: CVE-2026", fill="#FFE500")
d.text((95, 205), "[!] BYPASSING FIREWALL... [OK]", fill="#FF3B30")
d.text((95, 235), ">>> ACCESS GRANTED: ROOT", fill="#00FF66")
d.polygon([(30, 370), (482, 370), (500, 395), (12, 395)], fill="#1F2937", outline="#4B5563")
d.rectangle([210, 375, 302, 390], fill="#111827")
im.save(ASSET_DIR / "cyber-laptop.png")

# 2. Cyber Padlock Locked
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.arc([156, 70, 356, 290], start=180, end=0, fill="#00FF66", width=38)
d.rectangle([156, 170, 194, 250], fill="#00FF66")
d.rectangle([318, 170, 356, 250], fill="#00FF66")
d.rounded_rectangle([106, 210, 406, 450], radius=30, fill="#0A1910", outline="#00FF66", width=10)
d.ellipse([236, 280, 276, 320], fill="#00FF66")
d.polygon([(244, 310), (268, 310), (278, 380), (234, 380)], fill="#00FF66")
d.text((170, 400), "AES-256 SECURED", fill="#00FF66")
im.save(ASSET_DIR / "cyber-padlock-locked.png")

# 3. Cyber Padlock Unlocked / Broken
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.arc([196, 30, 396, 250], start=180, end=0, fill="#FF3B30", width=38)
d.rectangle([196, 130, 234, 250], fill="#FF3B30")
d.rectangle([358, 130, 396, 180], fill="#FF3B30")
d.rounded_rectangle([106, 210, 406, 450], radius=30, fill="#200A0A", outline="#FF3B30", width=10)
d.ellipse([236, 280, 276, 320], fill="#FF3B30")
d.polygon([(244, 310), (268, 310), (278, 380), (234, 380)], fill="#FF3B30")
d.text((160, 400), "PASSWORD CRACKED!", fill="#FF3B30")
im.save(ASSET_DIR / "cyber-padlock-unlocked.png")

# 4. Cyber Shield
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
pts = [(256, 50), (430, 90), (410, 320), (256, 460), (102, 320), (82, 90)]
d.polygon(pts, fill="#061A24", outline="#00E5FF", width=10)
d.line([(180, 260), (235, 320), (345, 180)], fill="#00FF66", width=22)
d.text((180, 370), "FIREWALL: ON", fill="#00E5FF")
im.save(ASSET_DIR / "cyber-shield.png")

# 5. Cyber Skull / Danger Alert
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.ellipse([136, 80, 376, 320], fill="#1E0B0B", outline="#FF3B30", width=10)
d.rounded_rectangle([186, 290, 326, 400], radius=16, fill="#1E0B0B", outline="#FF3B30", width=10)
d.ellipse([176, 170, 236, 240], fill="#FF3B30")
d.ellipse([276, 170, 336, 240], fill="#FF3B30")
d.polygon([(256, 260), (242, 290), (270, 290)], fill="#FF3B30")
for x in [216, 246, 276, 296]:
    d.line([(x, 320), (x, 380)], fill="#FF3B30", width=6)
d.text((150, 430), "CRITICAL THREAT!", fill="#FF3B30")
im.save(ASSET_DIR / "cyber-skull.png")

# 6. Database with SQL Injection Crack
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.ellipse([106, 60, 406, 150], fill="#1E293B", outline="#38BDF8", width=8)
d.rectangle([106, 105, 406, 380], fill="#0F172A")
d.line([(106, 105), (106, 380)], fill="#38BDF8", width=8)
d.line([(406, 105), (406, 380)], fill="#38BDF8", width=8)
d.arc([106, 160, 406, 250], start=0, end=180, fill="#38BDF8", width=8)
d.arc([106, 270, 406, 360], start=0, end=180, fill="#38BDF8", width=8)
d.ellipse([106, 340, 406, 430], fill="#1E293B", outline="#38BDF8", width=8)
d.line([(256, 180), (230, 230), (290, 280), (240, 350)], fill="#EF4444", width=8)
d.text((140, 200), "' OR 1=1 --", fill="#EF4444")
d.text((160, 450), "SQL INJECTION!", fill="#EF4444")
im.save(ASSET_DIR / "cyber-database.png")

# 7. Wi-Fi Hack Danger
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.ellipse([236, 380, 276, 420], fill="#F59E0B")
d.arc([186, 300, 326, 440], start=210, end=330, fill="#F59E0B", width=16)
d.arc([136, 220, 376, 460], start=210, end=330, fill="#F59E0B", width=16)
d.arc([86, 140, 426, 480], start=210, end=330, fill="#EF4444", width=20)
d.text((140, 90), "EVIL TWIN / ROGUE AP", fill="#EF4444")
d.text((170, 440), "UNSECURED WI-FI", fill="#F59E0B")
im.save(ASSET_DIR / "cyber-wifi.png")

# 8. Software Bug / Malware Icon
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.ellipse([156, 160, 356, 380], fill="#2B1010", outline="#EF4444", width=10)
d.ellipse([186, 100, 326, 210], fill="#2B1010", outline="#EF4444", width=8)
d.line([(220, 120), (170, 50)], fill="#EF4444", width=8)
d.line([(292, 120), (342, 50)], fill="#EF4444", width=8)
for y, dx in [(220, 110), (280, 120), (340, 100)]:
    d.line([(156, y), (156 - dx, y - 20), (156 - dx - 30, y + 30)], fill="#EF4444", width=8)
    d.line([(356, y), (356 + dx, y - 20), (356 + dx + 30, y + 30)], fill="#EF4444", width=8)
d.text((160, 420), "ZERO-DAY EXPLOIT", fill="#EF4444")
im.save(ASSET_DIR / "cyber-bug.png")

# 9. Terminal Window with Matrix Green Code
im = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.rounded_rectangle([40, 60, 472, 452], radius=16, fill="#0D1117", outline="#30363D", width=6)
# Window header
d.rounded_rectangle([40, 60, 472, 110], radius=16, fill="#161B22")
d.rectangle([40, 85, 472, 110], fill="#161B22")
d.ellipse([65, 78, 85, 98], fill="#FF5F56")
d.ellipse([95, 78, 115, 98], fill="#FFBD2E")
d.ellipse([125, 78, 145, 98], fill="#27C93F")
d.text((190, 77), "bash - 80x24", fill="#8B949E")
d.text((65, 130), "$ nmap -sV 192.168.1.1", fill="#00FF66")
d.text((65, 170), "Starting Nmap 7.94 scan...", fill="#E6EDF3")
d.text((65, 210), "PORT   STATE SERVICE VERSION", fill="#7EE787")
d.text((65, 250), "22/tcp open  ssh     OpenSSH 8.2", fill="#E6EDF3")
d.text((65, 290), "80/tcp open  http    Apache 2.4", fill="#E6EDF3")
d.text((65, 330), "[!] REVERSE SHELL OPENED!", fill="#FF7B72")
d.text((65, 370), "root@pwned:~# _", fill="#58A6FF")
im.save(ASSET_DIR / "cyber-terminal.png")

print("All 9 Cyber Props successfully generated in", ASSET_DIR)
