#!/usr/bin/env python3
"""Portable README contract checks, including original published demo assets."""
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
for name in ('first-use-guidance', 'b-common-quality', 'thumbnail-kit', 'docs'):
    subprocess.run([sys.executable, str(ROOT / 'tests' / f'verify-{name}.py')], cwd=ROOT, check=True)
for stem in ('light-theme-demo', 'dark-theme-demo', 'style-2a-demo', 'style-2b-demo',
             'style-3-demo', 'style-3-community-fire', 'style-4-cozy-station', 'style-4-cozy-cafe'):
    for extension in ('gif', 'mp4'):
        assert (ROOT / 'assets/readme' / f'{stem}.{extension}').is_file(), f'Missing demo: {stem}.{extension}'
for stem in ('style-5-japan', 'style-5-egypt'):
    assert (ROOT / 'assets/readme' / f'{stem}.mp4').is_file(), f'Missing historical demo: {stem}'
print('README contract passed.')
