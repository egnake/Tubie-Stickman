#!/usr/bin/env python3
"""Check local documentation links and the published catalog routes."""
from pathlib import Path
import re
from urllib.parse import unquote, urlsplit

ROOT = Path(__file__).resolve().parents[1]
paths = [ROOT / 'README.md', ROOT / 'examples/README.md',
         *sorted((ROOT / 'docs').glob('*.md')),
         *sorted((ROOT / 'docs/catalog').glob('*.md'))]
errors = []
for path in paths:
    content = path.read_text(encoding='utf-8')
    links = re.findall(r'\]\(([^)]+)\)', content)
    links += re.findall(r'(?:href|src)="([^"]+)"', content)
    for link in links:
        url = urlsplit(link)
        if url.scheme or url.netloc:
            continue
        target = path.parent / unquote(url.path) if url.path else path
        if not target.is_file():
            errors.append(f'{path.relative_to(ROOT)}: missing {link}')
            continue
        # Catalogs use explicit shared anchors for all navigation destinations.
        if url.fragment and target.parent == ROOT / 'docs/catalog':
            anchors = re.findall(r'<a id="([^"]+)"></a>', target.read_text(encoding='utf-8'))
            if unquote(url.fragment) not in anchors:
                errors.append(f'{path.relative_to(ROOT)}: missing anchor {link}')
assert not errors, '\n'.join(errors)
assert {p.name for p in (ROOT / 'docs/catalog').glob('README.*.md')} == {'README.en.md', 'README.tr.md'}
print(f'Documentation links passed ({len(paths)} files).')
