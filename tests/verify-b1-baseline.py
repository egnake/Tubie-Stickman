#!/usr/bin/env python3
"""Check the portable B1 acceptance evidence, not new-video visual quality."""
import hashlib
import json
import re
from pathlib import Path

skill = Path(__file__).resolve().parents[1] / "skills/directing-stickman-videos"
assets = skill / "assets/b1"
manifest = json.loads((assets / "approved-baseline.json").read_text())
assert manifest["schema"] == "b1-reference-evidence-v1"
assert manifest["style"] == "B1"
assert manifest["baseline_id"] == "b1-2026-10-23-ryan-v4"
assert {item["kind"] for item in manifest["evidence"]} == {
    "video_with_narration_captions_and_sfx", "contact_sheet", "encoded_motion_detail_sheet"
}

def in_skill(relative, base):
    assert not Path(relative).is_absolute(), "Reference must be portable"
    target = (base / relative).resolve()
    assert target.is_relative_to(skill.resolve()), "Reference escapes installed Skill"
    assert target.is_file(), f"Missing reference: {relative}"
    return target

standard = in_skill(manifest["production_standard"], assets)
for item in manifest["evidence"]:
    target = in_skill(item["path"], assets)
    assert hashlib.sha256(target.read_bytes()).hexdigest() == item["sha256"], target

doc = standard.read_text()
for relative in re.findall(r"\]\(([^)]+)\)", doc):
    if "://" not in relative:
        in_skill(relative, standard.parent)
entry = (skill / "SKILL.md").read_text()
assert "references/b1-minimal-handdrawn.md" in entry
assert manifest["narration_provenance"]["voice"] == "Bodi"
assert manifest["narration_provenance"]["speed"] == "native_unstretched"
assert "Ryan V4" in manifest["scope"]["public_showcase"]
assert "not GitHub publication" not in manifest["scope"]["public_showcase"]
assert "public preview remains V15" not in doc
license_text = (assets / "LUCIDE-LICENSE.txt").read_text()
assert "Lucide Contributors 2022" in license_text and "Permission to use, copy, modify" in license_text
assert "LUCIDE-LICENSE.txt" in (assets / "ATTRIBUTION.md").read_text()
print("B1 accepted baseline: three evidence fingerprints and all local reference links passed.")
print("Evidence is portable reference media, not a runtime project or proof of future output quality.")
