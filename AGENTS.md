# AGENTS.md

## Must-follow constraints
- **Do not edit hash-locked Director (`A1`–`A5`) or Voice contracts** without updating their SHA-256 baselines:
  - `skills/directing-stickman-videos/SKILL.md` from `## A · Director` to EOF is byte-locked in `tests/verify-ab-preservation.py` and `tests/verify-first-use-guidance.py`.
  - `skills/directing-stickman-videos/references/{style-catalog,prehistoric-dark-comedy,cozy-storytelling,historical-cartoon,voice-continuity,storyboard-template,omni-flash-prompt-contract,examples}.md` are SHA-256 locked across `tests/verify-ab-preservation.py`, `tests/verify-historical-cartoon.py`, and `tests/verify-voice-continuity.py`.
- **Keep `skills/directing-stickman-videos/` self-contained and relocatable**:
  - Tests copy `skills/directing-stickman-videos/` into `/tmp` without `node_modules` or the parent repo. All Markdown links, manifest paths, and `require()` paths inside the skill must stay relative and resolve strictly within `skills/directing-stickman-videos/` (no escaping `..` or external symlinks).
  - Never include `/Users/`, `/home/`, `/var/folders/`, `C:\`, or absolute paths in skill docs, manifests, or fixtures.
  - Keep `skills/directing-stickman-videos/SKILL.md` `<= 15000` bytes and keep it as the **only** `SKILL.md` in the repository.
  - Never add `preinstall`, `install`, or `postinstall` scripts to `skills/directing-stickman-videos/package.json`.
- **Do not modify or re-encode pinned media/baseline assets**:
  - Files in `skills/directing-stickman-videos/assets/{a6,b1,b3,b4,b5}/` (PNG/JPG references, MP4/AAC streams, `style-lock.txt`, and JSON manifests) and `assets/readme/thumbnails/*.jpg` (`1280x720`, `< 2MB`) are verified by exact SHA-256 and dimensions.
  - B5 scene scripts (`assets/b5/local-baselines/*/scene/*.cjs`) must remain deterministic and hermetic: do not use `child_process`, `process.env`, `fetch(`, or parent-repo paths.
- **Keep all tests 100% offline**:
  - Never add live TTS, Whisper model downloads, or paid API calls to tests. Use synthetic audio/video (`wave` or `ffmpeg -f lavfi`) and mocks.

## Validation before finishing
Run from the repository root:
```bash
python scripts/check.py
python tests/verify-readmes.py
python tests/verify-classic-character.py
python baslat.py --help
```

Shell/batch launchers were replaced by `baslat.py`; keep repository-owned `.sh` and `.bat` files absent.

## Repo-specific conventions
- **Editable install / `PYTHONPATH`**: `src/tubie_stickman/runtime.py` computes `ROOT_DIR` from `Path(__file__).resolve().parents[2]` to locate `skills/` and `assets/`. Always use editable installs (`pip install -e .`) or `PYTHONPATH=src`.
- **Node dependencies**: Workspace dependencies (`@napi-rs/canvas`, `sharp`) belong to `skills/directing-stickman-videos/package.json` and are installed from the repo root via `npm ci`.
- **Line endings**: `.gitattributes` enforces `LF` for `.py`, `.mjs`, `.md`, `.json`, `.yml`, `.html`, `.css`, and `.js`. Preserve `LF` on Markdown and text assets so byte-level SHA-256 checks pass on both Linux and Windows.

## Change safety rules
- **Catalog docs (`docs/catalog/`)**:
  - Only `README.en.md` and `README.tr.md` may exist matching `docs/catalog/README.*.md` (`tests/verify-docs.py` asserts the exact filename set).
  - Intra-catalog `#fragment` links require explicit `<a id="..."></a>` tags in the target file; Markdown heading slugs alone fail `tests/verify-docs.py`.
  - Preserve the `<!-- thumbnail-publishing-kit:start -->` ... `<!-- thumbnail-publishing-kit:end -->` block placement (between `<a id="b2"></a>` and `<a id="faq"></a>`), table structure (2 tables, 4 `<tr>`, 6 `<td align="center">`, zero `<th>`), and top nav link order enforced by `tests/verify-thumbnail-kit.py`.
- **Root `README*.md` restrictions**:
  - Never include internal A6 routing strings (`Route 2 · Renoise MCP (explicit choice)` or `utm_source=crazyegnake-aini-202610`) in root `README*.md` files (`tests/verify-a6-guidance.py`).
- **Untracked runtime artifacts**:
  - Never commit `sesim.wav`, `output/`, `projects/`, `narration.wav`, `bgm.wav`, `sfx.wav`, or `approval.json`.
