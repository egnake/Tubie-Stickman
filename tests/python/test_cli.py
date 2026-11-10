import unittest
from unittest.mock import patch

from tubie_stickman import cli


class CLITest(unittest.TestCase):
    def test_generate_dispatches_all_presets(self):
        with patch("tubie_stickman.video.CyberVideoPipeline") as pipeline:
            self.assertEqual(cli.main(["generate", "--all", "--mode", "pro"]), 0)
        pipeline.assert_called_once_with(mode="pro")
        self.assertEqual(pipeline.return_value.produce.call_count, 4)

    def test_record_opens_voice_gui(self):
        with patch("tubie_stickman.gui.main", return_value=0) as gui:
            self.assertEqual(cli.main(["record"]), 0)
        gui.assert_called_once_with(port=0, open_browser=True, tab="voice")

    def test_default_opens_gui_without_terminal_input(self):
        with patch("tubie_stickman.gui.main", return_value=0) as gui:
            self.assertEqual(cli.main([]), 0)
        gui.assert_called_once_with(port=0, open_browser=True, tab="produce")

    def test_gui_options(self):
        with patch("tubie_stickman.gui.main", return_value=0) as gui:
            self.assertEqual(cli.main(["gui", "--no-browser", "--port", "8765"]), 0)
        gui.assert_called_once_with(port=8765, open_browser=False, tab="produce")

    def test_conflicting_selection_is_rejected(self):
        with self.assertRaises(SystemExit):
            cli.main(["generate", "--all", "--preset", "1"])


if __name__ == "__main__":
    unittest.main()
