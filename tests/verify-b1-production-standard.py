#!/usr/bin/env python3
"""Portable B1 quality-evidence/link checks, not an animation/aesthetic evaluator."""
import hashlib
import json
import re
from pathlib import Path

repo = Path(__file__).resolve().parents[1]
skill = repo / "skills/directing-stickman-videos"
assets = skill / "assets/b1"
manifest_path = assets / "approved-production-standard.json"
manifest = json.loads(manifest_path.read_text())
assert manifest["schema"] == "b1-production-standard-v1"
assert manifest["style"] == "B1"
assert manifest["accepted_on"] == "2026-10-24"
assert manifest["scope"]["applies_to"].startswith("B1 only")
assert set(manifest["scope"]["not_inherited_by"]) == {
    "A1", "A2", "A3", "A4", "A5", "other B styles"
}
assert manifest["accepted_source"]["sha256"] == (
    "3739b0b1d8aef3678030cb4ac0c1d5ee0f36431f701a9ce0aad1adafb7f2e864"
)
assert manifest["accepted_source"]["frames"] == 5817
assert not (assets / manifest["accepted_source"]["file_name"]).exists(), "Do not bundle the local full film accidentally"

def local(relative, base):
    assert not Path(relative).is_absolute(), "Evidence links must be portable"
    target = (base / relative).resolve()
    assert target.is_relative_to(skill.resolve()), "Link escapes Skill"
    assert target.is_file(), f"Missing required file: {relative}"
    return target

standard = local(manifest["production_standard"], assets)
checklist = local(manifest["required_checklist"], assets)
common_checklist = local(manifest["common_checklist"], assets)
assert common_checklist.name == "autopilot-production-qa.md"
assert "not the independently shared B" in manifest["scope"]["not_inherited_scope"]
base = local(manifest["base_reference"], assets)
assert json.loads(base.read_text())["baseline_id"] == "b1-2026-10-23-ryan-v4"
assert len(manifest["evidence"]) == 1
for item in manifest["evidence"]:
    assert item["kind"] == "sampled_encoded_contact_sheet"
    target = local(item["path"], assets)
    assert hashlib.sha256(target.read_bytes()).hexdigest() == item["sha256"]

for doc in (standard, checklist):
    content = doc.read_text()
    assert "/Users/" not in content and "/var/folders/" not in content
    for relative in re.findall(r"\]\(([^)]+)\)", content):
        if "://" not in relative:
            local(relative, doc.parent)
assert "b1-production-qa.md" in standard.read_text(), "Checklist must be reachable from loaded B1 reference"
assert common_checklist.name in standard.read_text() and common_checklist.name in checklist.read_text(), "B1 must inherit B common quality separately from its style defaults"
assert list(skill.rglob("SKILL.md")) == [skill / "SKILL.md"], "Do not create another installable entrypoint"
assert (repo / "tests/scenarios/b1-production-quality.md").is_file()
assert "/Users/" not in manifest_path.read_text()
print("B1 production revision: portable links, accepted source/sheet provenance, original baseline and style isolation passed.")
print("Static integrity only: semantic, pacing, visual and listening quality require the production checklist on actual media.")
