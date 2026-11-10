"""Local-only Studio server, with isolated, cancellable production jobs."""
from __future__ import annotations

from collections import deque
import json
import mimetypes
import os
from pathlib import Path
import secrets
import signal
import subprocess
import sys
import tempfile
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import unquote, urlsplit
import webbrowser

from .runtime import ROOT_DIR
from .gui_worker import PREFIX

WEB = Path(__file__).with_name('web')
ACTIVE = {'running', 'approval', 'stopping'}
MAX_UPLOAD = 50 * 1024 * 1024


def media_path(name: str) -> Path:
    """Serve only generated deliverables, never arbitrary checkout files or symlinks."""
    relative = Path(name)
    if relative.is_absolute() or '..' in relative.parts or '\\' in name:
        raise ValueError('Geçersiz dosya yolu.')
    parts = relative.parts
    allowed = (len(parts) == 2 and parts[0] == 'output' and relative.suffix in {'.mp4', '.srt', '.png', '.wav'})
    allowed |= name == 'sesim.wav'
    allowed |= (len(parts) == 4 and parts[0] == 'projects' and parts[2] == 'output'
                and parts[3] in {'preview.mp4', 'live-frame.jpg'})
    candidate = ROOT_DIR / relative
    if not allowed or not candidate.resolve().is_relative_to(ROOT_DIR.resolve()):
        raise ValueError('Bu dosya paylaşılamaz.')
    if any(parent.is_symlink() for parent in [candidate, *candidate.parents] if parent != ROOT_DIR.parent):
        raise ValueError('Bağlantılı dosya paylaşılamaz.')
    return candidate


class Studio:
    def __init__(self):
        self.lock = threading.RLock()
        self.process = None
        self.logs = deque(maxlen=300)
        self.job = {'status': 'idle', 'stage': 'idle', 'label': 'Yeni bir üretime hazır', 'results': []}
        self.revision = 0

    def log(self, line):
        with self.lock:
            self.revision += 1
            self.logs.append({'id': self.revision, 'time': time.strftime('%H:%M:%S'), 'text': line[:4000]})

    def snapshot(self):
        with self.lock:
            outputs = []
            for file in sorted((ROOT_DIR / 'output').glob('*.mp4'), key=lambda p: p.stat().st_mtime, reverse=True):
                if not file.is_symlink():
                    outputs.append({'name': file.stem.replace('_', ' '), 'file': 'output/' + file.name,
                                    'size': file.stat().st_size,
                                    'subtitle': 'output/' + file.with_suffix('.srt').name if file.with_suffix('.srt').is_file() else None})
            return {'job': dict(self.job), 'logs': list(self.logs), 'outputs': outputs,
                    'reference': (ROOT_DIR / 'sesim.wav').is_file()}

    def start(self, action, arguments=(), cleanup=None):
        with self.lock:
            if self.job['status'] in ACTIVE:
                raise ValueError('Devam eden işlem tamamlanmalı veya iptal edilmeli.')
            self.logs.clear()
            self.job = {'status': 'running', 'stage': 'starting', 'label': 'İşlem başlatılıyor',
                        'action': action, 'started': time.time(), 'results': []}
            if action == 'generate' and len(arguments) >= 2:
                self.job.update(presets=arguments[0].split(','), mode=arguments[1])
            env = os.environ.copy()
            env['PYTHONPATH'] = str(ROOT_DIR / 'src') + (os.pathsep + env['PYTHONPATH'] if env.get('PYTHONPATH') else '')
            env['PYTHONIOENCODING'] = 'utf-8'
            env['PYTHONUNBUFFERED'] = '1'
            options = {'creationflags': subprocess.CREATE_NEW_PROCESS_GROUP} if os.name == 'nt' else {'start_new_session': True}
            try:
                self.process = subprocess.Popen(
                    [sys.executable, '-u', '-m', 'tubie_stickman.gui_worker', action, *arguments],
                    cwd=ROOT_DIR, env=env, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                    stderr=subprocess.STDOUT, text=True, encoding='utf-8', errors='replace', **options)
            except Exception:
                self.job.update(status='failed', label='İşlem başlatılamadı')
                if cleanup:
                    Path(cleanup).unlink(missing_ok=True)
                raise
            threading.Thread(target=self._read, args=(self.process, cleanup), daemon=True).start()

    def _read(self, process, cleanup):
        try:
            for line in process.stdout:
                line = line.rstrip()
                if not line:
                    continue
                if line.startswith(PREFIX):
                    try:
                        event = json.loads(line[len(PREFIX):])
                    except ValueError:
                        self.log(line)
                        continue
                    with self.lock:
                        if self.job['status'] == 'stopping':
                            continue
                        if 'result' in event:
                            self.job['results'].append(event['result'])
                        else:
                            if 'stage' in event:
                                self.job.pop('frame', None)
                                self.job.pop('frames', None)
                            if 'slug' in event:
                                for key in ('preview', 'live', 'frame', 'frames'):
                                    self.job.pop(key, None)
                            self.job.update(event)
                            if 'frame' in event and self.job.get('slug'):
                                self.job['live'] = f"projects/{self.job['slug']}/output/live-frame.jpg"
                                self.job['live_revision'] = time.time_ns()
                            if event.get('stage') == 'approval':
                                self.job['status'] = 'approval'
                else:
                    self.log(line)
            code = process.wait()
            with self.lock:
                cancelled = self.job['status'] == 'stopping'
                status = 'cancelled' if cancelled else 'complete' if code == 0 else 'failed'
                label = {'cancelled': 'İşlem iptal edildi', 'complete': 'İşlem tamamlandı',
                         'failed': 'İşlem tamamlanamadı · günlükteki ayrıntıyı kontrol edin'}[status]
                self.job.update(status=status, stage=status, label=label, ended=time.time())
                if code and not cancelled:
                    self.job['error'] = next((entry['text'] for entry in reversed(self.logs) if entry['text']), label)
                self.process = None
        finally:
            process.stdout.close()
            process.stdin.close()
            if cleanup:
                Path(cleanup).unlink(missing_ok=True)

    def approve(self):
        with self.lock:
            if self.job['status'] != 'approval' or not self.process or self.process.poll() is not None:
                raise ValueError('Onay bekleyen bir önizleme yok.')
            self.process.stdin.write('approve\n')
            self.process.stdin.flush()
            self.job.update(status='running', stage='render', label='Son video işleniyor')

    def cancel(self):
        with self.lock:
            process = self.process
            if not process or process.poll() is not None:
                return
            self.job.update(status='stopping', label='İşlem durduruluyor')
        if os.name == 'nt':
            subprocess.run(['taskkill', '/PID', str(process.pid), '/T', '/F'], capture_output=True, check=False)
        else:
            try:
                os.killpg(process.pid, signal.SIGTERM)
                process.wait(timeout=3)
            except subprocess.TimeoutExpired:
                pass
            except ProcessLookupError:
                return
            # The worker may exit while one of its descendants is still alive.
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass


class StudioServer(ThreadingHTTPServer):
    daemon_threads = True

    def __init__(self, address, studio):
        super().__init__(address, Handler)
        self.studio = studio
        self.token = secrets.token_urlsafe(24)
        self.prefix = '/' + self.token + '/'


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def route(self):
        # A capability URL plus loopback-only binding and exact Host/Origin checks.
        host = f'127.0.0.1:{self.server.server_port}'
        if self.headers.get('Host') != host:
            self.send_error(403)
            return None
        origin = self.headers.get('Origin')
        if origin and origin != 'http://' + host:
            self.send_error(403)
            return None
        path = unquote(urlsplit(self.path).path)
        if not path.startswith(self.server.prefix):
            self.send_error(404)
            return None
        return path[len(self.server.prefix):]

    def send_common_headers(self, status, mime, length):
        self.send_response(status)
        self.send_header('Content-Type', mime)
        self.send_header('Content-Length', str(length))
        self.send_header('Cache-Control', 'no-store')
        self.send_header('X-Content-Type-Options', 'nosniff')
        self.send_header('Referrer-Policy', 'no-referrer')
        self.send_header('Content-Security-Policy', "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; media-src 'self' blob:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'")

    def json(self, data, status=200):
        body = json.dumps(data, ensure_ascii=False).encode('utf-8')
        self.send_common_headers(status, 'application/json; charset=utf-8', len(body))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        route = self.route()
        if route is None:
            return
        try:
            if route == 'api/state':
                return self.json(self.server.studio.snapshot())
            if route == 'api/presets':
                from .video import CYBER_VIRAL_SCRIPTS
                return self.json([{'id': key, 'title': value['title'], 'narration': value['narration'],
                                   'scenes': len(value['scenes'])} for key, value in CYBER_VIRAL_SCRIPTS.items()])
            if route.startswith('media/'):
                path = media_path(route[6:])
            elif route in {'', 'index.html', 'app.css', 'app.js'}:
                path = WEB / (route or 'index.html')
            else:
                return self.send_error(404)
            self.file(path)
        except (ValueError, FileNotFoundError, IsADirectoryError):
            self.send_error(404)
        except (BrokenPipeError, ConnectionResetError):
            pass

    def file(self, path):
        with path.open('rb') as stream:
            size = os.fstat(stream.fileno()).st_size
            start, end, partial = 0, size - 1, False
            if self.headers.get('Range'):
                import re
                match = re.fullmatch(r'bytes=(\d*)-(\d*)', self.headers['Range'])
                if not match or not any(match.groups()):
                    return self.send_error(416)
                a, b = match.groups()
                if a:
                    start, end = int(a), min(int(b), size - 1) if b else size - 1
                else:
                    start = max(0, size - int(b))
                if start > end or start >= size:
                    self.send_response(416)
                    self.send_header('Content-Range', f'bytes */{size}')
                    self.send_header('Content-Length', '0')
                    self.end_headers()
                    return
                partial = True
            mime = mimetypes.guess_type(str(path))[0] or 'application/octet-stream'
            self.send_common_headers(206 if partial else 200, mime, max(0, end - start + 1))
            self.send_header('Accept-Ranges', 'bytes')
            if partial:
                self.send_header('Content-Range', f'bytes {start}-{end}/{size}')
            self.end_headers()
            stream.seek(start)
            remaining = end - start + 1
            while remaining > 0:
                block = stream.read(min(256 * 1024, remaining))
                if not block:
                    break
                self.wfile.write(block)
                remaining -= len(block)

    def do_POST(self):
        route = self.route()
        if route is None:
            return
        studio = self.server.studio
        try:
            length = int(self.headers.get('Content-Length', '0'))
            limit = MAX_UPLOAD if route == 'api/voice' else 4096
            if length < 0 or length > limit:
                return self.json({'error': 'Dosya çok büyük (en fazla 50 MB).'}, 413)
            body = self.rfile.read(length)
            if len(body) != length:
                raise ValueError('Eksik veri alındı.')
            if route == 'api/voice':
                if not body:
                    raise ValueError('Bir ses dosyası seçin.')
                with studio.lock:
                    if studio.job['status'] in ACTIVE:
                        raise ValueError('Önce devam eden işlemi tamamlayın.')
                    with tempfile.NamedTemporaryFile(prefix='tubie-upload-', suffix='.audio', delete=False) as upload:
                        upload.write(body)
                    studio.start('voice', [upload.name], cleanup=upload.name)
            elif route == 'api/job':
                data = json.loads(body or b'{}')
                if not isinstance(data, dict):
                    raise ValueError('Geçersiz seçim.')
                action = data.get('action', 'generate')
                if action == 'generate':
                    keys = data.get('presets', [])
                    mode = data.get('mode', 'pro')
                    if (not isinstance(keys, list) or not keys or len(keys) > 4
                            or any(key not in ('1', '2', '3', '4') for key in keys)
                            or len(set(keys)) != len(keys) or mode not in ('pro', 'cloned', 'auto')):
                        raise ValueError('Şablon ve ses seçimini kontrol edin.')
                    if mode == 'cloned' and not (ROOT_DIR / 'sesim.wav').is_file():
                        raise ValueError('Önce Ses sekmesinden referans ses yükleyin veya kaydedin.')
                    studio.start(action, [','.join(keys), mode])
                elif action in ('doctor', 'test-voice'):
                    studio.start(action)
                else:
                    raise ValueError('Geçersiz işlem.')
            elif route == 'api/approve':
                studio.approve()
            elif route == 'api/cancel':
                studio.cancel()
            elif route == 'api/shutdown':
                studio.cancel()
                threading.Thread(target=self.server.shutdown, daemon=True).start()
            else:
                return self.json({'error': 'İşlem bulunamadı.'}, 404)
            self.json({'ok': True})
        except (ValueError, TypeError, OSError) as exc:
            self.json({'error': str(exc)}, 400)
        except (BrokenPipeError, ConnectionResetError):
            pass


def main(port=0, open_browser=True, tab='produce'):
    studio = Studio()
    with StudioServer(('127.0.0.1', port), studio) as server:
        url = f'http://127.0.0.1:{server.server_port}{server.prefix}#{tab}'
        print(f'Tubie Studio: {url}', flush=True)
        print('Kapatmak için GUI’deki Çıkış düğmesini veya Ctrl+C kullanın.', flush=True)
        if open_browser:
            try:
                if not webbrowser.open(url):
                    print('Tarayıcı otomatik açılamadı; yukarıdaki adresi açın.', flush=True)
            except webbrowser.Error:
                print('Tarayıcı açılamadı; yukarıdaki adresi açın.', flush=True)
        try:
            server.serve_forever(poll_interval=0.25)
        except KeyboardInterrupt:
            pass
        finally:
            studio.cancel()
    return 0
