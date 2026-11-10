"""Portable runtime behavior that does not require media or ML dependencies."""
import sys
import unittest
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "src"))

from tubie_stickman import runtime, video, voice


class RuntimeTests(unittest.TestCase):
    def test_resources_use_checkout_root(self):
        root = Path(__file__).resolve().parents[2]
        self.assertEqual(runtime.ROOT_DIR, root)
        self.assertEqual(video.SKILL_DIR, root / "skills" / "directing-stickman-videos")
        self.assertEqual(voice.SESIM_WAV, root / "sesim.wav")

    def test_binary_path_prefers_path(self):
        with patch.object(runtime.shutil, "which", return_value="/usr/bin/node"):
            self.assertEqual(runtime.get_binary_path("node"), "/usr/bin/node")

    def test_cloned_mode_requires_reference(self):
        with patch.object(video, "SESIM_WAV", Path("/nonexistent/tubie-voice.wav")):
            with self.assertRaisesRegex(FileNotFoundError, "referans"):
                video.VoiceEngine(mode="cloned").synthesize("Merhaba", Path("/tmp/tubie-test.wav"))

class AudioConversionTests(unittest.TestCase):
    def test_reference_conversion_accepts_same_source_and_preserves_on_failure(self):
        import tempfile
        import subprocess
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory) / "sesim.wav"
            target.write_bytes(b"original")
            with patch.object(voice, "SESIM_WAV", target):
                with patch.object(voice.subprocess, "run", side_effect=subprocess.CalledProcessError(1, "ffmpeg")):
                    with self.assertRaises(subprocess.CalledProcessError):
                        voice._convert_reference(target)
                self.assertEqual(target.read_bytes(), b"original")
                def convert(command, **kwargs):
                    self.assertEqual(Path(command[command.index("-i") + 1]), target)
                    self.assertNotEqual(Path(command[-1]), target)
                    Path(command[-1]).write_bytes(b"converted")
                with patch.object(voice.subprocess, "run", side_effect=convert):
                    voice._convert_reference(target)
                self.assertEqual(target.read_bytes(), b"converted")
                self.assertEqual(list(Path(directory).iterdir()), [target])

    def test_edge_speech_is_converted_to_wav(self):
        import tempfile
        import types
        async def save(path):
            Path(path).write_bytes(b"mp3")
        fake_edge = types.SimpleNamespace(Communicate=lambda *a, **kw: types.SimpleNamespace(save=save))
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory) / "narration.wav"
            def convert(command, **kwargs):
                self.assertEqual(Path(command[command.index("-i") + 1]).read_bytes(), b"mp3")
                self.assertEqual(command[-5:-1], ["-ar", "24000", "-ac", "1"])
                Path(command[-1]).write_bytes(b"wav")
            with patch.dict(sys.modules, {"edge_tts": fake_edge}), patch.object(video.subprocess, "run", side_effect=convert), patch.object(video.VoiceEngine, "_get_duration", return_value=1):
                video.VoiceEngine(mode="pro").synthesize("Merhaba", target)
            self.assertEqual(target.read_bytes(), b"wav")

    def test_environment_override(self):
        with patch.dict("os.environ", {"FFMPEG": "/custom/ffmpeg"}):
            self.assertEqual(runtime.get_binary_path("ffmpeg"), "/custom/ffmpeg")


if __name__ == "__main__":
    unittest.main()
