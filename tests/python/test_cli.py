import unittest
from unittest.mock import patch

from tubie_stickman import cli


class CLITest(unittest.TestCase):
    def test_generate_dispatches_all_presets(self):
        with patch("tubie_stickman.video.CyberVideoPipeline") as pipeline:
            self.assertEqual(cli.main(["generate", "--all", "--mode", "pro"]), 0)
        pipeline.assert_called_once_with(mode="pro")
        self.assertEqual(pipeline.return_value.produce.call_count, 4)

    def test_record_dispatches(self):
        with patch("tubie_stickman.voice.main", return_value=0) as record:
            self.assertEqual(cli.main(["record"]), 0)
        record.assert_called_once_with()

    def test_record_failure_propagates(self):
        with patch("tubie_stickman.voice.main", return_value=1):
            self.assertEqual(cli.main(["record"]), 1)

    def test_interactive_exit(self):
        with patch("builtins.input", return_value="0"):
            self.assertEqual(cli.main([]), 0)

    def test_conflicting_selection_is_rejected(self):
        with self.assertRaises(SystemExit):
            cli.main(["generate", "--all", "--preset", "1"])


if __name__ == "__main__":
    unittest.main()
