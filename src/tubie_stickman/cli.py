"""Command line entry point for the video and voice tools."""
from __future__ import annotations

import argparse
import importlib.util
import shutil
import sys
from pathlib import Path

from .runtime import ROOT_DIR, get_binary_path


def parser() -> argparse.ArgumentParser:
    root = argparse.ArgumentParser(prog="tubie", description="Stickman video üretimi")
    commands = root.add_subparsers(dest="command")
    generate = commands.add_parser("generate", help="Hazır şablonlardan video üret")
    selection = generate.add_mutually_exclusive_group()
    selection.add_argument("--preset", choices=["1", "2", "3", "4"], help="Şablon numarası (varsayılan: 1)")
    selection.add_argument("--all", action="store_true", help="Tüm şablonları üret")
    generate.add_argument("--mode", choices=["auto", "pro", "cloned"], default="auto", help="Ses modu")
    commands.add_parser("record", help="Ses kaydet veya ses dosyası yükle")
    commands.add_parser("doctor", help="Ortam bağımlılıklarını kontrol et")
    return root


def doctor() -> int:
    """Report required tools and optional voice features without loading heavy models."""
    missing = False
    for name in ("node", "ffmpeg", "ffprobe"):
        candidate = get_binary_path(name)
        path = candidate if Path(candidate).is_file() else shutil.which(candidate)
        print(f"{'OK' if path else 'MISSING'} {name}: {path or 'PATH üzerinde bulunamadı'}")
        missing |= path is None
    for module, required, hint in (
        ("edge_tts", True, "pip install -e ."),
        ("faster_whisper", True, "pip install -e ."),
        ("sounddevice", False, "pip install -e '.[record]'"),
        ("soundfile", False, "pip install -e '.[record]'"),
        ("TTS", False, "pip install -e '.[clone]'"),
        ("torch", False, "pip install -e '.[clone]'"),
    ):
        present = importlib.util.find_spec(module) is not None
        print(f"{'OK' if present else ('MISSING' if required else 'OPTIONAL')} {module}" + ("" if present else f": {hint}"))
        missing |= required and not present
    # The Node runtime is maintained in the skill's own package and lockfile.
    skill = ROOT_DIR / "skills" / "directing-stickman-videos"
    for module in ("@napi-rs/canvas", "sharp"):
        present = any((base / "node_modules" / module).is_dir() for base in (skill, ROOT_DIR))
        print(f"{'OK' if present else 'MISSING'} node module {module}" + ("" if present else f": npm ci --prefix {skill}"))
        missing |= not present
    return int(missing)


def interactive() -> int:
    print("\nStickman Video Director\n  [1] Video üret\n  [2] Ses kaydet / yükle\n  [3] Ortamı kontrol et\n  [0] Çıkış")
    choice = input("Seçiminiz [1]: ").strip() or "1"
    if choice == "0":
        return 0
    if choice == "2":
        return main(["record"])
    if choice == "3":
        return main(["doctor"])
    if choice != "1":
        print("Geçersiz seçim.", file=sys.stderr)
        return 2
    print("\n[1] Wi-Fi  [2] Brute force  [3] SQL injection  [4] Phishing  [A] Tümü")
    selection = (input("Şablon [1]: ").strip() or "1").lower()
    if selection not in {"1", "2", "3", "4", "a"}:
        print("Geçersiz şablon.", file=sys.stderr)
        return 2
    mode = (input("Ses modu: [1] Profesyonel  [2] Klonlanmış [1]: ").strip() or "1")
    if mode not in {"1", "2"}:
        print("Geçersiz ses modu.", file=sys.stderr)
        return 2
    return main(["generate", *( ["--all"] if selection == "a" else ["--preset", selection]), "--mode", "cloned" if mode == "2" else "pro"])


def main(argv: list[str] | None = None) -> int:
    args = parser().parse_args(argv)
    if args.command is None:
        return interactive()
    if args.command == "doctor":
        return doctor()
    if args.command == "record":
        from . import voice
        return voice.main() or 0
    from . import video
    pipeline = video.CyberVideoPipeline(mode=args.mode)
    for key in (["1", "2", "3", "4"] if args.all else [args.preset or "1"]):
        pipeline.produce(video.CYBER_VIRAL_SCRIPTS[key])
    return 0
