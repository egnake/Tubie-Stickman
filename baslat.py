#!/usr/bin/env python3
"""Windows ve Linux için ortak Tubie Studio başlatıcısı."""
from pathlib import Path
import os
import sys


def main():
    root = Path(__file__).resolve().parent
    python = root / '.venv' / ('Scripts/python.exe' if os.name == 'nt' else 'bin/python')
    if python.is_file() and Path(sys.prefix).resolve() != python.parent.parent.resolve():
        os.execv(str(python), [str(python), str(Path(__file__).resolve()), *sys.argv[1:]])
    sys.path.insert(0, str(root / 'src'))
    from tubie_stickman.cli import main as run
    return run()


if __name__ == '__main__':
    raise SystemExit(main())
