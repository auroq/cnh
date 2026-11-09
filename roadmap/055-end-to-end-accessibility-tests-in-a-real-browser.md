# 55. End-to-end accessibility tests in a real browser, run by the CD gate

**Done 2026-11-09 (PR #65, issue #55). Filed 2026-11-09 by Aria. Nobody's call: the target users define it, and the checks are the ones `.gitar/review/a11y.md` already asks for. The `cd.yml` wiring is the integrator's, from the snippet in the PR.**

Depends on: 010, 036

Numbered 055 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

`testing/e2e/` is a Playwright and TypeScript project with project-local dependencies only, no global installs.
It drives Chromium against `WIDEDOOR_URL` and runs `@axe-core/playwright` on the pages.
It proves what a screen reader, keyboard and large-text user actually gets, which the Go unit tests and the HTTP acceptance tests cannot.

## Why

Merge Forward "Conferences for Everyone": the app exists for attendees with limited mobility, hearing or vision.
A regression that breaks keyboard focus or contrast is a product failure, and it should fail the deploy.

## Acceptance

- An axe scan of the home page and of an answer page (ask "nearest all-gender restroom" from "255 BC") reports zero violations at WCAG 2.2 AA, and the axe JSON report is saved as a CI artifact.
- Keyboard-only flow: Tab reaches the skip link, then the question field, the location field, the Ask button, then each quick pick in order; Enter on Ask submits; after the answer page loads the `aria-live` region contains the summary text and the answer heading is the first thing after the skip-link target.
- Every quick pick link's accessible name equals its visible text; the page has exactly one `h1` and a `lang` attribute.
- At 200% zoom (viewport 640x900 with deviceScaleFactor 2) and at 320px width there is no horizontal scroll and the Ask button is within the viewport.
- With `prefers-color-scheme: dark` and with `prefers-contrast: more` emulated, axe colour contrast passes.
- `make test-e2e` runs it locally and installs Chromium into the project cache on first run; tests are idempotent and parallel-safe.
- Wiring into `cd.yml` after the Rollout is Healthy and before the acceptance step, with a failure taking the existing revert path, is done by the integrator session from a ready-to-paste job snippet in the PR body.

## Touches

`testing/e2e/` (package.json, playwright.config.ts, helpers, web specs), `Makefile`, `.gitignore`.
Tests follow `.gitar/review/testing.md` naming in `test.describe("when ...")`, `test.describe("and ...")` and `test("it should ...")` form, one expectation per test.
