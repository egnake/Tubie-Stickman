#!/usr/bin/env python3
"""Build lightweight README previews; preserve all original media."""
from pathlib import Path
import shutil
import subprocess

ROOT = Path(__file__).resolve().parents[1]

def main():
    ffmpeg = shutil.which('ffmpeg')
    if not ffmpeg:
        raise SystemExit('ffmpeg is required.')
    media = ROOT / 'assets/readme'
    (media / 'previews').mkdir(exist_ok=True)
    original = ('light-theme-demo', 'dark-theme-demo', 'style-2a-demo', 'style-2b-demo',
                'style-3-demo', 'style-3-community-fire', 'style-4-cozy-station', 'style-4-cozy-cafe')
    for stem, extension, fps, colors, scale, width in (
        *[(stem, 'gif', 8, 96, 4, 400) for stem in original],
        ('style-5-japan', 'mp4', 6, 64, 5, 400),
        ('style-5-egypt', 'mp4', 6, 64, 5, 400),
        ('style-5-japan', 'mp4', 6, 64, 5, 640),
    ):
        output = f'{stem}-hero.gif' if width == 640 else f'{stem}.gif'
        filters = f'[0:v]fps={fps},scale={width}:{width * 9 // 16}:flags=lanczos,split[a][b];[a]palettegen=max_colors={colors}:stats_mode=diff[p];[b][p]paletteuse=dither=bayer:bayer_scale={scale}:diff_mode=rectangle'
        command = [ffmpeg, '-hide_banner', '-loglevel', 'error', '-y', '-i', str(media / f'{stem}.{extension}')]
        if extension == 'mp4':
            command += ['-t', '10']
        subprocess.run([*command, '-filter_complex', filters, '-loop', '0', str(media / 'previews' / output)], check=True)
    print('Built lightweight README previews; original media unchanged.')

if __name__ == '__main__':
    main()
