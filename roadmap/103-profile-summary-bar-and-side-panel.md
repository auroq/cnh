# 103. The profile is a summary bar under the header on every page, a side panel open by default on wide screens, and the full panel on `/profile`

**Open. Filed 2026-11-09 by Aria (issue #103). Aria's call: she set the layout.**

Depends on: 081, 085

Numbered 103 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and Aria's correction are on [issue #103](https://github.com/auroq/cnh/issues/103); this file is the summary the roadmap keeps.

## What

"What Wide Door knows about you" leaves the collapsed `<details>`.
A one-line summary bar sits directly under the header and before `<main>` in DOM order on every page: "What Wide Door knows about you: 2 things. Edit".
The full panel is an `<aside aria-labelledby="profile-heading">` after `<main>` in DOM order, holding the heading, the chips with Add, Remove and Keep, Forget me, Save my profile and the disclosure sentence.
On wide screens (min-width 60rem) the aside is a sticky right column, open, and the bar's Edit link points at `#profile`.
On narrow screens the aside is not shown to anyone, sighted or not, and the bar's Edit link goes to `/profile`, which renders the full aside under the heading at both widths.
No JavaScript is needed for any of it.

## Why

Aria's direction: on a screen big enough the profile should be visible and open by default, because seeing what the app knows is the point; on a phone it should be reachable in one tap without pushing the answer below the first screen.

## Acceptance

- At 1280px width the aside is visible beside the answer without interaction and the bar's Edit link targets `#profile`.
- At 390px width the summary bar is within the first viewport height, the aside is not shown, and the Edit link goes to `/profile`.
- `/profile` shows the full aside at both widths.
- DOM order is skip link, header, bar, main, aside, footer; axe passes at both widths; one `h1`.
- Keyboard: at 1280px Tab reaches the aside's controls after the answer; at 390px Tab reaches the Edit link.
- Playwright in `testing/e2e/web/profile.spec.ts` covers both layouts; the zoom cases still pass without horizontal scroll.
- Unit tests in `pkg/web` cover the rendered structure.

## Touches

`pkg/web/templates/home.html`, `pkg/web/static/style.css`, `pkg/web/profile.go`, `pkg/web/server.go`, `pkg/web/profile_test.go`, `testing/e2e/web/profile.spec.ts`, `testing/e2e/web/keyboard.spec.ts`.
