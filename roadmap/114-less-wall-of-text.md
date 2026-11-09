# 114. Less wall of text: one visual language for answers, notes, warnings, verification and controls

**Done 2026-11-09 (issue #114). Aria's direction: the eye and the screen reader both find the answer first, then what matters about it, then what they can do.**

Depends on: 096, 103

Numbered 114 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text is on [issue #114](https://github.com/auroq/cnh/issues/114); this file is the summary the roadmap keeps.

## What

The answer sentence carries the `t-answer` class and is the largest text after the page heading, with space above and below.
Callouts share one shape: `note` (grey), `info` (blue, "Published by the organiser"), `warning` (amber, triangle glyph, "Unverified" and disputed reports), `success` (green, check glyph, "Verified by attendees"), plus the red `problem` pair for reported problems and errors.
Each callout starts with a text label, visible or visually hidden, so a screen reader hears "Warning:" or "Verified:" before the body and colour is never the only signal.
Place and session results are cards with a border, route facts as compact chips, and the "Were you here?" disclosure styled as an outlined 44 px control.
Four type sizes only: heading 2rem, answer 1.5rem, body 1rem, meta 0.9rem, with prose capped at 70 characters.
Dark and `prefers-contrast: more` variants keep every pair at or above 4.5:1; high contrast drops the fills and keeps borders and labels.

## Why

The page carries a lot of text.
A sighted or low-vision user on a phone should see the answer, then what matters about it, then what they can do, without reading everything.

## Acceptance

- Playwright suite (axe at 390 and 1280, dark, `prefers-contrast: more`, keyboard order, zoom) passes unchanged.
- Each callout has a text label as its first content.
- The summary sentence is the only element with the `t-answer` class.
- No markup reordering, no new interactive behaviour, no JavaScript.
