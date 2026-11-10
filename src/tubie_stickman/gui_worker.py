"""Isolated GUI jobs. stdout carries logs and structured progress; never shell code."""
from __future__ import annotations

import json
from pathlib import Path
import sys

PREFIX = 'TUBIE_EVENT '


def emit(**event):
    print(PREFIX + json.dumps(event, ensure_ascii=False), flush=True)


def main():
    action = sys.argv[1]
    if action == 'doctor':
        from .cli import doctor
        return doctor()
    if action == 'voice':
        from .voice import _convert_reference
        source = Path(sys.argv[2])
        try:
            emit(stage='voice', label='Referans ses hazırlanıyor')
            _convert_reference(source)
        finally:
            source.unlink(missing_ok=True)
        emit(stage='complete', label='Referans ses hazır')
        return 0
    if action == 'test-voice':
        from .voice import ses_testi_yap
        if not ses_testi_yap():
            return 1
        emit(audio='output/klonlanmis_ses_testi.wav')
        return 0
    from .video import CYBER_VIRAL_SCRIPTS, CyberVideoPipeline
    keys, mode = sys.argv[2].split(','), sys.argv[3]
    def review(project):
        from .runtime import ROOT_DIR
        emit(stage='approval', label='Önizleme hazır · onayınızı bekliyor',
             preview=(project / 'output/preview.mp4').relative_to(ROOT_DIR).as_posix())
        if sys.stdin.readline().strip() != 'approve':
            raise RuntimeError('Önizleme onaylanmadı.')
    pipeline = CyberVideoPipeline(mode=mode, reporter=lambda **data: emit(**data), review=review)
    for index, key in enumerate(keys):
        emit(stage='voice', label='Seslendirme hazırlanıyor', title=CYBER_VIRAL_SCRIPTS[key]['title'],
             item=index + 1, items=len(keys), slug=CYBER_VIRAL_SCRIPTS[key]['slug'])
        output = pipeline.produce(CYBER_VIRAL_SCRIPTS[key])
        from .runtime import ROOT_DIR
        emit(result=output.relative_to(ROOT_DIR).as_posix())
    return 0


if __name__ == '__main__':
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f'Hata: {exc}', file=sys.stderr, flush=True)
        raise SystemExit(1)
