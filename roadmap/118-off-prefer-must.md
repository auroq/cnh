# 118. Each need is Off, Prefer or Must: Must filters, Prefer ranks, and the profile says which

**Open (issue #118). Aria's direction: "I would rather avoid stairs" and "I cannot do stairs" are different facts.**

Depends on: 103, 114

Numbered 118 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text is on [issue #118](https://github.com/auroq/cnh/issues/118); this file is the summary the roadmap keeps.

## What

Every chip carries a strength: Prefer or Must, encoded as `step-free` and `step-free!` in `prefs` and the `wd_prefs` cookie.
A learned chip is always Prefer; keeping it makes it a chosen Prefer, never a Must.
Must removes every place that does not meet it.
When nothing remains, the nearest place stays, marked "does not meet", and the answer says "Nothing we know meets: step-free. Nearest otherwise: Main entrance, stairs."
Prefer ranks as before.
The API exposes `profile.applied[].strength`.

## Done so far

Model, parsing, the filter versus rank behaviour per chip, the sentence, the API field, with unit tests.

## Left

The `/profile` radio groups (Off, Prefer, Must) and the bar and aside wording.
Acceptance rows for `step-free!` and the e2e test for the radio group.
