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
    commands.add_parser("record", help="GUI ses merkezini aç")
    gui = commands.add_parser("gui", help="Tubie Studio arayüzünü aç")
    gui.add_argument("--port", type=int, default=0, help="Yerel port (varsayılan: otomatik)")
    gui.add_argument("--no-browser", action="store_true", help="Tarayıcıyı otomatik açma")
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


def main(argv: list[str] | None = None) -> int:
    args = parser().parse_args(argv)
    if args.command in (None, "gui", "record"):
        from . import gui
        return gui.main(port=getattr(args, "port", 0),
                        open_browser=not getattr(args, "no_browser", False),
                        tab="voice" if args.command == "record" else "produce")
    if args.command == "doctor":
        return doctor()
    from . import video
    pipeline = video.CyberVideoPipeline(mode=args.mode)
    for key in (["1", "2", "3", "4"] if args.all else [args.preset or "1"]):
        pipeline.produce(video.CYBER_VIRAL_SCRIPTS[key])
    return 0
