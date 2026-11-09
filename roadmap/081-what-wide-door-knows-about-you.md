# 81. What Wide Door knows about you: an anonymous profile of chips, on your device, shaping every answer

**Done 2026-11-09 (PR #93, issue #81). Filed 2026-11-09 by Aria, p1, from a judge's suggestion. Aria's call: the chip list and the wording; the server stays stateless about people. Left for a follow-up: the QR code on Save my profile, and the learned chips of item 085, whose source the encoding already carries.**

Depends on: 063

Numbered 081 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #81](https://github.com/auroq/cnh/issues/81); this file is the summary the roadmap keeps.

## What

A profile of need chips (step-free, avoid crowds and noise, quiet spaces matter, all-gender restrooms, captions or ASL, and so on) shown on every page under "What Wide Door knows about you", stored only in the browser, sent with each question, used to shape the answer, with the sentence "Stored only on this device. Sent with your question and forgotten. We never ask what you have, only what you need."
Chips are buttons with accessible names, operable by keyboard, working without JavaScript through a form POST to `/profile`; `/api/ask` reports `profile.applied`.
The server sets no cookie of its own and keeps no profile in any index or log.

## Why

A judge's suggestion: the step up from "ask anything" is answers shaped by the person, done ethically, with what we know shown on the screen.
Merge Forward: chips are needs, not diagnoses.

## Acceptance

- Adding and removing a chip works with keyboard only and with JavaScript off.
- The same question with and without the step-free chip ranks a step-free entrance first only with it.
- No request leaves a profile in the server's logs or indexes.

## Touches

`pkg/model`, `pkg/web`, `pkg/api/v1`, `pkg/ask`.
