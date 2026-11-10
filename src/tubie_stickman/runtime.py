"""Runtime paths and external program discovery for the source checkout."""
import os
import shutil
import sys
from pathlib import Path

ROOT_DIR = Path(__file__).resolve().parents[2]


def get_binary_path(name: str) -> str:
    """Find an executable on PATH, with WinGet's link directory as a fallback."""
    override = os.environ.get(name.upper())
    if override:
        return override
    found = shutil.which(name)
    if found:
        return found
    if sys.platform == "win32" and os.environ.get("LOCALAPPDATA"):
        candidate = Path(os.environ["LOCALAPPDATA"]) / "Microsoft" / "WinGet" / "Links" / f"{name}.exe"
        if candidate.is_file():
            return str(candidate)
    return name


def cuda_device() -> str:
    """Use CPU when PyTorch is absent or CUDA is unavailable."""
    try:
        import torch
    except ImportError:
        return "cpu"
    return "cuda" if torch.cuda.is_available() else "cpu"
