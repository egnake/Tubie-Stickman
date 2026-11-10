#!/usr/bin/env python3
"""Style 5 routing and source-anchor regression; not a provider quality score."""
from pathlib import Path
import hashlib

root = Path(__file__).resolve().parents[1]
skill = root / "skills/directing-stickman-videos"
ref = (skill / "references/historical-cartoon.md").read_text()
anchor = ref.split("```text\n", 1)[1].split("```", 1)[0]
# Exact CHARACTER/FACE/PALMS/EXPOSED LIMBS from the owner-selected prompt.
# Fingerprint verifies source preservation, not reliable generated anatomy.
assert hashlib.sha256(anchor.encode()).hexdigest() == "f7f8161688c3c17a4379eaf5e4d64d7c0888f7e4b94ef59b7beb3dcafc785a8e"
for token in (
    "https://github.com/leksautomate", "/pull/41",
    "repeat its **full description**", "two visible non-color differences",
    "no style-specific character-count cap", "voice-continuity.md",
    "default when no other voice is requested", "opening/action/ending states",
    "Backgrounds are full-color", "do not rewrite Styles 1–4",
    "black mitten-palm design with integrated thumb lobe and wrist-to-sleeve attachment",
    "not sleeves or footwear", "Grips, releases and gestures may change pose",
    "This clarification is A5-only",
    "compact black mitten palms and continuous wrist attachment during walking, grips, releases and empty-hand gestures",
    "no individual fingers, finger separations, white fills or gloves",
    "contact or overlap between the palm and the prop",
    "without spreading fingers",
    "NO INDIVIDUAL FINGERS, NO FINGER SEPARATIONS, NO WHITE HANDS OR GLOVES",
    "not frozen poses or permanently clenched fists",
    "Earlier round-tip or exposed-neck tests do not validate this revised silhouette",
    "short, sturdy white neck", "one continuous filled silhouette",
    "about one tenth of the head diameter", "roughly four head-heights tall",
    "character-specific eyebrows, nose and mouth", "18% of head diameter long and 10% wide",
    "stickman animated video", "smaller than half an eye's height",
    "On A5 selection", "front/side/back", "every clip in a supporting tool",
    "Users can knowingly choose text-only", "not evidence of reliable text-only generation",
    "rather than offering that advice only when asked", "only list specific reference files when the user actually supplies them",
    "generate their own front/side/back", "This is advice only",
    "do not supply or automatically generate reference images",
    "This connection rule is A5-only",
    "continuous white jaw–short-neck–garment connection and any beard/collar overlap",
    "A background-colored gap fails",
    "The owner reference defines the target, not cross-clip stability",
    "not overall stability passes",
):
    assert token in ref, token
for obsolete in (
    "Small white-filled, black-outlined cartoon hands attach naturally",
    "Hands are small simplified white-filled shapes",
    "small white-filled black-outlined hands connected at the wrists",
    "with minimal finger strokes or curved grips",
    "A few short finger strokes can indicate an open hand",
    "A short thin black neck line connects the base of each head to the neckline",
    "with the same thin black line treatment as the exposed limbs",
    "A short white-filled, black-outlined neck smoothly connects the base of each head to the neckline",
    "Show a clearly visible short white neck",
    "one-eighth of the head diameter",
    "must not conceal the entire white connection",
    "Do not use a black stick neck, fuse the head into the shoulders or lengthen the neck",
    "mark the neck check failed or unverified rather than passed",
    "The head joins the clothed body naturally at the neckline; no separately visible neck is required.",
    "Hands are small solid black round or oval tips",
    "NO THUMBS",
    "The base of the head may meet or overlap the garment neckline directly",
    "a short collar connection or fully concealed neck is acceptable",
    "No separate neck is exposed", "no nose, and thin black stick limbs",
    "The lower edge of the white round head physically touches",
    "about 1.5-2 wrist-stroke widths", "Keep palm length below one sixth",
    "Reference images remain optional when requested and supported",
):
    assert obsolete not in ref, f"Obsolete A5 anatomy instruction: {obsolete}"
for filename in (
    "SKILL.md", "references/style-catalog.md",
    "references/storyboard-template.md", "references/omni-flash-prompt-contract.md",
):
    assert "historical-cartoon.md" in (skill / filename).read_text(), filename

for filename in (
    "SKILL.md", "references/style-catalog.md", "references/omni-flash-prompt-contract.md",
):
    text = (skill / filename).read_text()
    a5_line = next(line for line in text.splitlines() if "Style 5" in line)
    if filename == "SKILL.md":
        assert "compact fingerless palms" in a5_line, filename
        assert "connected short white necks" in a5_line, filename
    else:
        assert "small solid black mitten-shaped palms continuous with the wrists" in a5_line, filename
        assert "one subtle integrated thumb lobe, no individual fingers, finger separations, white fills or gloves" in a5_line, filename
        assert "connected short white necks" in a5_line, filename
    assert "white-filled black-outlined cartoon hands" not in a5_line, filename

# Existing specialist contracts must not be rewritten for this addition.
baseline = {
    "prehistoric-dark-comedy.md": "1a3a58a22b392cc087a0aa645b3f8478a3136e7e7f35d086ddb588ac091df860",
    "cozy-storytelling.md": "a0b3f5890857582bb1e7bfa819cefcc7ffd55af72ab8997e1b286b7ad27f1419",
    "voice-continuity.md": "24014bc599e237644c743f033a0b62c556ba4ed8689d382d08b1bcdce1967de2",
}
for filename, fingerprint in baseline.items():
    path = skill / "references" / filename
    content = path.read_bytes()
    if filename == "voice-continuity.md":
        # Only the exact A/B scope notice may precede the unchanged A contract.
        heading = b"# Shared voice continuity \xe2\x80\x94 all visual styles\n\n"
        notice = (
            b"Scope: **A \xc2\xb7 Director text-to-video prompts only**. B \xc2\xb7 Autopilot uses named TTS voices under `autopilot.md`; "
            b"do not copy this document's descriptive voice profiles into any B-style TTS request.\n\n"
        )
        assert content.startswith(heading + notice), "Missing exact A-only voice scope"
        content = heading + content[len(heading + notice):]
    assert hashlib.sha256(content).hexdigest() == fingerprint, filename
print("Style 5 source anchor, routing, scope and unchanged specialist contracts passed.")
