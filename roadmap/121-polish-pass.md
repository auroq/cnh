# 121. One polish pass: clean, calm, pretty, and still accessible

**Done 2026-11-09 (issue #121). Aria's direction: feature complete, so one pass so the demo looks as considered as it is.**

Depends on: 114

Numbered 121 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text is on [issue #121](https://github.com/auroq/cnh/issues/121); this file is the summary the roadmap keeps.

## What

Spacing moves to 0.5rem steps everywhere: headings, fields, cards, callouts, aside and footer.
One primary action per screen: Ask is the only solid button, quick picks and Forget me are outlined, Keep and the "Were you here?" disclosure are text buttons with an underline on hover and a border on focus, and a chip that is on carries a light accent tint instead of a solid fill.
The header is a neutral band with an inline door mark (SVG, `aria-hidden`) beside the text wordmark, the lede as a one line tagline, and the profile bar sitting inside the band.
Cards share one shape: 1rem padding, a 1px border, a 0.5rem radius, a surface colour that has a visible edge in the dark scheme, the name as the only bold line and meta in muted text.
The ask form puts both inputs and the button on one baseline from 48rem up and stacks them below that.
The error callout gets the card shape.
The footer is compact.
`prefers-contrast: more` drops the band, surface and tint fills and restores borders on the text buttons.

## Why

Rhythm, hierarchy and one primary action help low vision and cognitive load users as much as anyone,
and the demo should look as considered as the accessibility work underneath it.

## Acceptance

- Playwright suite (axe at 390 and 1280, light, dark, `prefers-contrast: more`) passes, run by the CD gate.
- Diff is `style.css`, two class attributes and the header SVG; no markup reordering, no JavaScript, no change to landmarks, Tab order, live regions or e2e selectors.
- Every colour pair touched is at or above 4.5:1 for text and 3:1 for borders.
