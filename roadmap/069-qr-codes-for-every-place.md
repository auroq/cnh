# 69. QR codes for every room and landmark: scan to tell Wide Door where you are

**Open. Filed 2026-11-09 by Aria (issue #69). Nobody's call: the URL shape, the kinds that get a code and the sheet are all read off the places data that already exists.**

Depends on: 008, 010

Numbered 069 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

Every room, entrance and landmark in the places index has a URL that pre-fills the location: `https://widedoor.cnh.auroq.com/?from=<place id>`.
`widedoor qr` renders one QR code per place (SVG and PNG) into an output directory, and with `--sheet` a printable `sheet.html` with one page per place: the place name in large print, the building and level, the URL in text under the code, so a volunteer can print and tape it beside the door sign.
The page, when opened from a QR link, says where it thinks you are ("You scanned the code at Grand Ballroom J, Salt Palace Level 1") and lets you change it; the location field is pre-filled with the room name and focus lands on the question field, so a phone user asks in one step.
`Locate` tries a place id first, because ids are unambiguous, and falls back to the room search when the id is unknown.

## Why

Merge Forward: this app is used on a phone, standing in a corridor.
Scanning the code on the door is faster and more accurate than typing a room name, and it works for a blind attendee whose phone reads the code aloud.
It also closes the loop on "where am I" for people who cannot read the signage.

## Acceptance

- `widedoor qr --out <dir>` writes `<place id>.svg` and `<place id>.png` for every place of kind room, entrance, landmark, quiet-room, prayer-room, nursing-room, first-aid, info-desk or coat-check; the QR encodes the production URL with `from=<place id>`; the library is pure Go (`github.com/skip2/go-qrcode`) and the SVG is rendered by hand from its bitmap.
- `--sheet` also writes `sheet.html`: one page per place, print CSS, the place name at 48pt minimum, the URL in text under the code, the building and level.
- `/?from=<id>` resolves the id to the place and the page shows "You scanned the code at <name>, <venue> <level>" with a change link; the question field has `autofocus` only when the page was opened from a code.
- `/api/ask` accepts a place id in `from` as well as free text; ids are tried first in `Locate`.
- Unit tests for the URL building, the sheet and the id-first resolution; an acceptance row that `/?from=room-salt-palace-level-1-grand-ballroom-j` renders the scanned-at sentence.
- A note in the README under Run it on how to print the sheet.

## Touches

`pkg/qr` (new adapter), `pkg/cli` (qr subcommand), `pkg/config` (`qr.base_url`), `pkg/ask` (id-first Locate, small), `pkg/esindex` (id lookup), `pkg/web` (scanned-at wording, autofocus rule), `testing/acceptance`, `testing/helpers`, `tools/archcheck.sh`, `docs/architecture.md`, README.

## Sequencing

Issue #63 is changing `pkg/ask` at the same time, so the qr subcommand, the sheet and the `pkg/web` wording land first and the id-first `Locate` branch is the last, smallest commit.
