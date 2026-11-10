"""Offline GUI transport and worker lifecycle tests; no synthesis or model downloads."""
import http.client
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import threading
import time
import unittest
from unittest.mock import patch

from tubie_stickman import gui


def until(predicate, timeout=8):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if predicate():
            return
        time.sleep(.02)
    raise AssertionError('Worker did not reach expected state')


class WorkerTests(unittest.TestCase):
    def setUp(self):
        self.studio = gui.Studio()
        self.spawn = subprocess.Popen

    def tearDown(self):
        self.studio.cancel()
        until(lambda: self.studio.process is None)

    def worker(self, code):
        def spawn(command, **kwargs):
            return self.spawn([sys.executable, '-u', '-c', code], **kwargs)
        return patch.object(gui.subprocess, 'Popen', side_effect=spawn)

    def test_preview_approval_logs_and_completion(self):
        code = """
import json, sys
print('model log')
print('TUBIE_EVENT ' + json.dumps({'stage': 'approval', 'preview': 'projects/test/output/preview.mp4'}))
assert sys.stdin.readline().strip() == 'approve'
print('TUBIE_EVENT ' + json.dumps({'result': 'output/test.mp4'}))
"""
        with self.worker(code):
            self.studio.start('generate', ['2', 'pro'])
        until(lambda: self.studio.job['status'] == 'approval')
        with self.assertRaises(ValueError):
            self.studio.start('doctor')
        self.assertEqual(self.studio.snapshot()['logs'][0]['text'], 'model log')
        self.assertEqual(self.studio.job['presets'], ['2'])
        self.assertEqual(self.studio.job['mode'], 'pro')
        self.studio.approve()
        until(lambda: self.studio.job['status'] == 'complete')
        self.assertEqual(self.studio.job['results'], ['output/test.mp4'])
        with self.assertRaises(ValueError):
            self.studio.approve()

    def test_cancel_waiting_worker_then_restart(self):
        with self.worker("import sys; print('TUBIE_EVENT {\"stage\":\"approval\"}'); sys.stdin.readline()"):
            self.studio.start('generate')
        until(lambda: self.studio.job['status'] == 'approval')
        self.studio.cancel()
        until(lambda: self.studio.job['status'] == 'cancelled')
        with self.worker("print('restart works')"):
            self.studio.start('doctor')
        until(lambda: self.studio.job['status'] == 'complete')

    def test_failure_and_log_bound(self):
        with self.worker("import sys; [print('line', i) for i in range(350)]; print('failure detail'); sys.exit(7)"):
            self.studio.start('doctor')
        until(lambda: self.studio.job['status'] == 'failed')
        self.assertEqual(self.studio.job['error'], 'failure detail')
        self.assertEqual(len(self.studio.logs), 300)

    @unittest.skipIf(os.name == 'nt', 'POSIX process-group behavior')
    def test_cancel_kills_descendants(self):
        code = "import subprocess, sys, time; p=subprocess.Popen([sys.executable, '-c', 'import time; time.sleep(120)']); print(p.pid); time.sleep(120)"
        with self.worker(code):
            self.studio.start('generate')
        until(lambda: len(self.studio.logs) > 0)
        pid = int(self.studio.logs[0]['text'])
        self.studio.cancel()
        until(lambda: self.studio.process is None)
        def stopped():
            status = Path(f'/proc/{pid}/stat')
            return not status.exists() or status.read_text().split()[2] == 'Z'
        until(stopped)


class TransportTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.root = Path(self.directory.name)
        self.root_patch = patch.object(gui, 'ROOT_DIR', self.root)
        self.root_patch.start()
        self.studio = gui.Studio()
        try:
            self.server = gui.StudioServer(('127.0.0.1', 0), self.studio)
        except PermissionError:
            self.root_patch.stop()
            self.directory.cleanup()
            self.skipTest('Loopback sockets unavailable in this sandbox')
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()
        self.prefix = self.server.prefix

    def tearDown(self):
        self.server.shutdown()
        self.server.server_close()
        self.thread.join()
        self.root_patch.stop()
        self.directory.cleanup()

    def request(self, path, method='GET', data=None, headers=None):
        connection = http.client.HTTPConnection('127.0.0.1', self.server.server_port, timeout=5)
        body = data if isinstance(data, bytes) else json.dumps(data) if data is not None else None
        connection.request(method, path, body=body, headers=headers or {})
        response = connection.getresponse()
        result = response.status, dict(response.getheaders()), response.read()
        connection.close()
        return result

    def test_page_capability_and_host_origin_checks(self):
        self.assertEqual(self.request('/')[0], 404)
        status, headers, body = self.request(self.prefix)
        self.assertEqual(status, 200)
        self.assertIn(b'Tubie Studio', body)
        self.assertIn("frame-ancestors 'none'", headers['Content-Security-Policy'])
        self.assertEqual(self.request(self.prefix, headers={'Host': 'malicious.example'})[0], 403)
        self.assertEqual(self.request(self.prefix + 'api/job', 'POST', {}, {'Origin': 'https://malicious.example'})[0], 403)

    def test_video_range_and_traversal(self):
        (self.root / 'output').mkdir()
        (self.root / 'output/test.mp4').write_bytes(b'0123456789')
        status, headers, body = self.request(self.prefix + 'media/output/test.mp4', headers={'Range': 'bytes=2-5'})
        self.assertEqual((status, body, headers['Content-Range']), (206, b'2345', 'bytes 2-5/10'))
        self.assertEqual(self.request(self.prefix + 'media/output/test.mp4', headers={'Range': 'bytes=99-'})[0], 416)
        self.assertEqual(self.request(self.prefix + 'media/output/test.mp4', headers={'Range': 'bytes=-3'})[2], b'789')
        for name in ('../secret', 'output/../../README.md', 'output/%2e%2e/README.md', 'output/%5csecret.mp4'):
            self.assertEqual(self.request(self.prefix + 'media/' + name)[0], 404)
        if os.name != 'nt':
            (self.root / 'output/link.mp4').symlink_to(Path(__file__))
            self.assertEqual(self.request(self.prefix + 'media/output/link.mp4')[0], 404)

    def test_voice_upload_is_private_and_bounded(self):
        with patch.object(self.studio, 'start') as start:
            status, _, _ = self.request(self.prefix + 'api/voice', 'POST', b'RIFF-test-audio')
        self.assertEqual(status, 200)
        source = Path(start.call_args.args[1][0])
        try:
            self.assertEqual(source.read_bytes(), b'RIFF-test-audio')
            self.assertEqual(start.call_args.kwargs['cleanup'], str(source))
        finally:
            source.unlink()
        self.assertEqual(self.request(self.prefix + 'api/voice', 'POST', b'')[0], 400)
        self.assertEqual(self.request(self.prefix + 'api/voice', 'POST', b'',
                                     {'Content-Length': str(gui.MAX_UPLOAD + 1)})[0], 413)
        self.studio.job['status'] = 'running'
        self.assertEqual(self.request(self.prefix + 'api/voice', 'POST', b'audio')[0], 400)
        self.studio.job['status'] = 'idle'
        (self.root / 'sesim.wav').write_bytes(b'private-reference')
        self.assertEqual(self.request(self.prefix + 'media/sesim.wav')[2], b'private-reference')
        self.assertEqual(self.request('/media/sesim.wav')[0], 404)

    def test_start_validation_and_voice_reference_required(self):
        for data in ({'presets': []}, {'presets': ['1', '1']}, {'presets': ['5']}, {'presets': ['1'], 'mode': 'other'}, {'presets': ['1'], 'mode': 'cloned'}, []):
            self.assertEqual(self.request(self.prefix + 'api/job', 'POST', data)[0], 400)
        with patch.object(self.studio, 'start') as start:
            self.assertEqual(self.request(self.prefix + 'api/job', 'POST', {'presets': ['1'], 'mode': 'pro'})[0], 200)
        start.assert_called_once_with('generate', ['1', 'pro'])
        self.assertEqual(json.loads(self.request(self.prefix + 'api/state')[2])['job']['status'], 'idle')


if __name__ == '__main__':
    unittest.main()
