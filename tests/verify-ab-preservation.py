#!/usr/bin/env python3
"""Pre-B byte baseline, with owner-directed A5 hand and neck clarifications.

This checks preserved instructions, not provider-backed visual/audio quality.
The voice reference baseline includes the pre-existing local A-only scope notice.
Only the explicit A5 changes below are allowed; A1–A4 remain byte-identical.
"""
import hashlib
from pathlib import Path

skill = Path(__file__).resolve().parents[1] / "skills/directing-stickman-videos"
expected = {
    "style-catalog.md": "0c75bdb6fefab28ac6f0dc61f6fa151a0793667479e23ae2696f4964fd4ea01d",
    "prehistoric-dark-comedy.md": "1a3a58a22b392cc087a0aa645b3f8478a3136e7e7f35d086ddb588ac091df860",
    "cozy-storytelling.md": "a0b3f5890857582bb1e7bfa819cefcc7ffd55af72ab8997e1b286b7ad27f1419",
    "historical-cartoon.md": "360f374c7b521353f5fbb9cb546b6f9f70a098973b99281c28d04ce28e77b01d",
    "voice-continuity.md": "ac14fce4f8e4176815d05260de8eeea22345c0920d8e3e2b3911a380affe2ed3",
    "storyboard-template.md": "486338e79397c73443fb390ff03883e5712834b1ddf542c3af54a8dbf6cb870b",
    "omni-flash-prompt-contract.md": "2ef4527750d10c20274f89f86cca9d6483430cef489722c4afb9893ae1d8fbe5",
    "examples.md": "c183e159304eb6dc12f9ad0b06f037a3f880282bc0b75bf06a2863705ddf00f4"
}
# Exact owner-directed A5 baseline revision (2026-10-25). Normalize only
# these three A5 summary lines to the pre-B baseline; all other bytes stay frozen.
a5_summary_refinements = {
    "SKILL.md": [
        "  - `Style 5 (Historical Cartoon / 历史卡通风)`: white-filled round heads, simple black oval eyes and thin black stick limbs in era-specific costume, with cel shading and full-color painterly historical settings",
        "  - `Style 5 (Historical Cartoon / 历史卡通风)`: white heads with connected short white necks, expressive faces, black stick limbs and compact fingerless palms, period dress and cel-shaded historical settings"
    ],
    "style-catalog.md": [
        "| **Style 5** | Historical Cartoon (历史卡通风) | Full-color painterly historical settings, clean cel shading, era-specific wardrobe and props | History explainers, historical daily life, civilization stories and historical comedy | White filled round heads, simple black oval eyes, thick outlines and thin black stick limbs |",
        "| **Style 5** | Historical Cartoon (历史卡通风) | Full-color painterly historical settings, clean cel shading, era-specific wardrobe and props | History explainers, historical daily life, civilization stories and historical comedy | White filled round heads with connected short white necks, expressive individual faces, solid black oval eyes, thin black exposed limbs and small solid black mitten-shaped palms continuous with the wrists, one subtle integrated thumb lobe, no individual fingers, finger separations, white fills or gloves |"
    ],
    "omni-flash-prompt-contract.md": [
        "Style 5 (Historical Cartoon / 历史卡通风): read `historical-cartoon.md` for its standalone prompt order and overrides. Repeat the full frozen drawing anchor, visible cast cards, approved era pack and complete `voice-continuity.md` voice block in every clip. Keep solid oval eyes and black stick limbs, with historical costume and colored painterly scenery. Do not inherit Style 1's hollow-head/no-face/no-clothes contract or generic piano, three-color, four-device and no-idle requirements. Follow its era score, physical beats and planned cuts; do not shorten approved anchors to meet an arbitrary character budget.",
        "Style 5 (Historical Cartoon / 历史卡通风): read `historical-cartoon.md` for its standalone prompt order and overrides. Repeat the full frozen drawing anchor, visible cast cards, approved era pack and complete `voice-continuity.md` voice block in every clip. Keep connected short white necks, expressive individual faces, solid oval eyes, thin black exposed limbs and small solid black mitten-shaped palms continuous with the wrists, one subtle integrated thumb lobe, no individual fingers, finger separations, white fills or gloves, with historical costume and colored painterly scenery. Do not inherit Style 1's hollow-head/no-face/no-clothes contract or generic piano, three-color, four-device and no-idle requirements. Follow its era score, physical beats and planned cuts; do not shorten approved anchors to meet an arbitrary character budget."
    ]
}


def pre_b_bytes(name, content):
    if name in a5_summary_refinements:
        old, new = (text.encode() for text in a5_summary_refinements[name])
        assert content.count(new) == 1, f"Missing exact A5 refinement: {name}"
        content = content.replace(new, old, 1)
    return content


# The dedicated A5 reference has its own exact approved patch fingerprint.
# Do not refresh the shared-file baseline or weaken A1–A4/voice comparisons.
# 2026-10-25 owner-selected rich-face/short-white-neck baseline and proactive
# three-view guidance, narrowed to user-generated-reference advice only.
# This fingerprint is scope evidence, not visual approval.
a5_contact_mitten_reference_sha256 = "74ad606ea2d8a237806661b0ee5e81ead4759aae003cee50bf328f2b383a38bd"

entry = pre_b_bytes("SKILL.md", (skill / "SKILL.md").read_bytes())
director = b"## A \xc2\xb7 Director\n" + entry.split(b"## A \xc2\xb7 Director\n", 1)[1]
assert hashlib.sha256(director).hexdigest() == "f710612be73c952fb130d891b0650388ec5ae89469cae95eda8e147eec18b73e", "A Director contract changed"
for name, digest in expected.items():
    content = (skill / "references" / name).read_bytes()
    if name == "historical-cartoon.md":
        assert hashlib.sha256(content).hexdigest() == a5_contact_mitten_reference_sha256, "Unreviewed A5 reference change"
    else:
        assert hashlib.sha256(pre_b_bytes(name, content)).hexdigest() == digest, f"A reference changed: {name}"
print("A1–A4, shared prompts and voice retain the pre-B bytes; only exact A5 baseline refinements are allowed.")
print("No live generation or visual-quality guarantee is implied.")
