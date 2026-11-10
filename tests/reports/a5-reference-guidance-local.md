# A5 character-reference guidance — local maintenance check

Date: 2026-10-25. Scope: existing `directing-stickman-videos`, A5 only. No new generation, paid calls, global installation or publication in this update.

## Changes

- Route A5 selection to a proactive three-view recommendation before planning, in the conversation language.
- Recommend confirmed front/side/back views, reuse approved images, and attach approved reference files alongside each complete clip prompt in a supporting model/mode.
- Keep text-only available without a new mandatory gate or repeated questions. References supplement prompts and are not a guarantee.
- Replace the superseded no-exposed-neck/no-nose recipe with the exact four character blocks from the owner's selected prompt: connected short white neck, adult proportions, rich individual face, compact black fingerless palms and black exposed limbs. Do not promote failed V31 collar experiments as a stable production solution.
- Keep A1–A4, B production and shared voice rules unchanged.

## Checks

PASS: `verify-first-use-guidance.py`, `verify-historical-cartoon.py`, `verify-ab-preservation.py`, `verify-b-common-quality.py`, `verify-voice-continuity.py`, `verify-cozy-storytelling.py`, and `git diff --check`.

PASS: exact comparison of CHARACTER/FACE/PALMS/EXPOSED LIMBS to the owner-supplied source. Anchor SHA-256: `f7f8161688c3c17a4379eaf5e4d64d7c0888f7e4b94ef59b7beb3dcafc785a8e`.

Root entry remains 13,997 bytes. Static checks verify instructions and scope, not model behavior or generated-video quality. Three new A5 conversational acceptance cases are authored but not model-run.

The generic Qiaomu package validator did not pass: this legacy package lacks its required `README.md`, `agents/interface.yaml` and `manifest.json` layout. No unrelated package migration was performed. The generic quick validator could not run under the current Python because PyYAML is absent; no dependencies were installed solely for that check.

## Evidence boundary

Design advantage: proactive reference guidance addresses the user's workflow concern. Missing evidence: fresh-host conversational run, a new reference-conditioned video batch, and clean-install verification for this revision. Recent text-only tests still included France head/neck disconnection and Egypt missing required hair; neither the revised prompt nor three-view guidance proves stable video generation.

This is a local source update only. Remote GitHub and any separately installed copy remain unchanged.
