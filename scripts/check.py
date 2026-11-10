#!/usr/bin/env python3
"""Run local regression checks on Windows and Linux without external generation."""
from pathlib import Path
import os
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
env = os.environ.copy()
env['PYTHONPATH'] = str(ROOT / 'src') + (os.pathsep + env['PYTHONPATH'] if env.get('PYTHONPATH') else '')
env['PYTHONIOENCODING'] = 'utf-8'
env['PYTHONUTF8'] = '1'
checks = [
    [sys.executable, '-m', 'unittest', 'discover', '-s', 'tests/python', '-v'],
    [sys.executable, '-m', 'unittest', 'discover', '-s', 'skills/directing-stickman-videos/tests', '-p', 'test_*.py', '-v'],
]
checks += [[sys.executable, str(path.relative_to(ROOT))] for path in sorted((ROOT / 'tests').glob('verify-*.py'))]
checks.append([shutil.which('npm') or 'npm', 'test'])
failed = []
for command in checks:
    print('\nRunning:', ' '.join(command), flush=True)
    result = subprocess.run(command, cwd=ROOT, env=env, check=False)
    if result.returncode:
        failed.append(command)
if failed:
    print(f'\n{len(failed)} check group(s) failed.', file=sys.stderr)
    raise SystemExit(1)
print('\nAll checks passed.')
