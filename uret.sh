#!/bin/sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$ROOT"
if [ -x "$ROOT/.venv/bin/python" ]; then PYTHON="$ROOT/.venv/bin/python"; else PYTHON=python3; fi
PYTHONPATH="$ROOT/src${PYTHONPATH:+:$PYTHONPATH}" exec "$PYTHON" -m tubie_stickman "$@"
