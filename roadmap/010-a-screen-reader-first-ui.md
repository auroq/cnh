# 10. A screen-reader-first UI

**Done 2026-11-09 (PR #19 and #22, issue #10). Filed 2026-11-09, from the kickoff plan. Nobody's call: the target users define it.**

Depends on: 009

## What

A single page served by the API: a "where are you" room picker, a list of feature buttons (elevator, quiet room, all-gender restroom, step-free entrance, captioned session) and one question box.
Semantic HTML with no custom widgets where a native one exists, answers announced through an ARIA live region, every action reachable by keyboard with a visible focus ring.
Text scales to 200% without horizontal scroll, colours meet WCAG AA contrast, and `prefers-contrast: more` and `forced-colors` are honoured.
No map image is required to get an answer; a map, if there is time, is an extra.

## Why

Vision, mobility and hearing are the Merge Forward focus, and the research is explicit that attendees ask for information, not a map they cannot read.
A blind attendee has to be able to use it without help, or the whole build misses its point.

## Acceptance

- With a screen reader and keyboard only, asking for the nearest quiet room from room 255 reads the answer aloud without moving focus by hand.
- An automated accessibility check (axe or equivalent) reports no violations on the page.
- At 200% browser zoom on a 390px-wide viewport nothing is cut off.

## Touches

`api/` static assets or `ui/`, depending on what item 001 chose.
