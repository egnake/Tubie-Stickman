"""Exercise the Python-to-Node render path using synthetic local audio."""
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest
from unittest.mock import patch
import wave

from tubie_stickman import video


@unittest.skipUnless(shutil.which('node') and shutil.which('ffmpeg'), 'Node and FFmpeg required')
class PipelineTest(unittest.TestCase):
    def test_portrait_render_from_unicode_directory(self):
        root = video.BASE_DIR
        if not any((base / 'node_modules/@napi-rs/canvas').is_dir() for base in (root, video.SKILL_DIR)):
            self.skipTest('Run npm ci first')
        with tempfile.TemporaryDirectory(prefix="tubie Türkçe video's-") as name:
            temp = Path(name)
            pipeline = video.CyberVideoPipeline(mode='pro')

            def synthesize(text, destination):
                with wave.open(str(destination), 'wb') as audio:
                    audio.setnchannels(1)
                    audio.setsampwidth(2)
                    audio.setframerate(24000)
                    audio.writeframes(b'\x00\x00' * 24000)
                return 1.0, 'pro'

            script = {
                'title': 'Türkçe test', 'slug': 'turkce_test', 'narration': 'Merhaba dünya.',
                'scenes': [video.CYBER_VIRAL_SCRIPTS['1']['scenes'][0]],
            }
            with patch.object(video, 'OUTPUT_DIR', temp / 'output'), \
                 patch.object(video, 'PROJECTS_DIR', temp / 'projects'), \
                 patch.object(pipeline.voice_engine, 'synthesize', side_effect=synthesize), \
                 patch.object(pipeline.aligner, 'align', return_value=[
                     {'word': 'Merhaba', 'start': 0, 'end': 0.4},
                     {'word': 'dünya', 'start': 0.4, 'end': 0.9},
                 ]):
                film = pipeline.produce(script)
            self.assertGreater(film.stat().st_size, 1000)
            self.assertTrue(film.with_suffix('.srt').is_file())
            result = subprocess.run([
                video.FFPROBE_EXE, '-v', 'error', '-show_streams', '-of', 'json', str(film)
            ], check=True, capture_output=True, text=True, encoding='utf-8')
            streams = json.loads(result.stdout)['streams']
            picture = next(s for s in streams if s['codec_type'] == 'video')
            self.assertEqual((picture['width'], picture['height']), (720, 1280))
            self.assertTrue(any(s['codec_type'] == 'audio' for s in streams))
