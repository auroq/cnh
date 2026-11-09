# 112. Chips name needs, not people: "I use a wheelchair or scooter" becomes "Wide doors and ramps"

**Done 2026-11-09 (PR #NN, issue #112). Aria's call: chips name needs, not people.**

Depends on: 085, 103

Numbered 112 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text is on [issue #112](https://github.com/auroq/cnh/issues/112); this file is the summary the roadmap keeps.

## What

The `wheels` chip's label changes from "I use a wheelchair or scooter" to "Wide doors and ramps" everywhere it is shown: the profile aside, the Add, Remove and Keep accessible names, the live region sentences and the learned-chip sentence.
The chip id `wheels` stays, so cookies and saved links set before the rename keep working; the README notes that the id is historical.
The learned mapping from item 085 still fires on "wheelchair" and "scooter" and on questions understood as mobility equipment; its sentence becomes "Added Wide doors and ramps to what Wide Door knows about you, because you asked about a wheelchair or scooter."
The other six labels were checked against the same rule and pass: "Step-free only", "Avoid crowds and noise", "Quiet spaces matter", "All-gender restrooms", "Captions or ASL", "Dietary needs, unspecified".

## Why

The profile promises to ask only what you need, never what you have.
A chip that names a device and a person breaks that promise,
and it makes a wheelchair user, a scooter user, a parent with a stroller and a person with a walker self-identify to get the same door.
Disability style guidance, the W3C WAI material on writing about people with disabilities and most national guides, says to describe the barrier or the requirement, not the user.

## Acceptance

- The string "I use a wheelchair or scooter" appears nowhere in `pkg/`, `testing/`, `README.md` or `docs/`.
- Unit tests and e2e specs that referenced the old label pass with the new one.
- The learned sentence for "nearest wheelchair ramp" names "Wide doors and ramps".

## Touches

`pkg/model/preferences.go`, `pkg/ask/learn.go`, `pkg/ask/learn_test.go`, `pkg/web/learn_test.go`, `testing/acceptance/profile_test.go`, `testing/e2e/web/profile.spec.ts`, `README.md`.
