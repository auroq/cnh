# 2. The app needs a name

**Done 2026-11-09. Filed 2026-11-09, from the kickoff plan. Aria's call: the app is Wide Door, the team is Curb Cut.**

Depends on: nothing

## What

The app is **Wide Door**: the thing an attendee is looking for, in words that need no accessibility vocabulary.
The team is **Curb Cut**, after the curb-cut effect: access built for disabled people ends up used by everyone.
"Nearby" and "Ask the venue" were the placeholders; "Ask the venue" stays as the name of Option 1 in the ADRs.
The name goes in the README title, the UI's page title and the hostname under `cnh.auroq.com`.
The repository stays `auroq/cnh` whatever the name is.

## Why

A screen reader announces the page title first, so the name is the first thing a blind attendee hears (Merge Forward).
Judges for Best overall see it before anything else.

## Acceptance

- The README title, `<title>` and the HTTPRoute hostname all use the chosen name.

## Touches

`README.md`, the UI's HTML head, the app's HTTPRoute.
