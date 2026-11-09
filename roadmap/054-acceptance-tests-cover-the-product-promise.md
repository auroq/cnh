# 54. Acceptance tests cover the product promise, as a table of question, place and expected answer

**Open. Filed 2026-11-09 by Aria (issue #54). Nobody's call: the rows are read off the live app and the organiser's page, nothing in them is a design choice.**

Depends on: 009, 036

Numbered 054 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

The acceptance tests in `testing/acceptance/` read from one table of cases: the question, where the attendee is, the expected top place, the expected distance wording (same level, same building, or no distance because the room is unknown or not given) and whether the organiser published it.
Every quick pick on the home page has at least one row, there is a row from a room on each level of each building, and there are rows with no room given and with a room nobody has heard of.
The JSON answer's shape (`summary`, `here{asked,room,venue,level}`, `places[]`, `sessions[]`) and the `/healthz` and `/readyz` probes are checked in the same package.

## Why

Merge Forward and Elasticsearch tracks: the table is the product promise a judge can read in one screen, and the CD gate holds the deployed app to it on every deploy.

## Acceptance

- `testing/acceptance/cases.go` holds the table; each row runs as nested `t.Run` blocks per `.gitar/review/testing.md`, one assertion per leaf (top place, wording, verified).
- The starter rows from the issue are all present and pass against https://widedoor.cnh.auroq.com.
- `make test-acceptance` runs it against `WIDEDOOR_URL`; the CD gate already runs `./testing/acceptance/...`, so no workflow changes.
- A quick pick that answers the wrong kind from some room is reported on issue #54, not written into the table as the expectation.

## Touches

`testing/acceptance/`, `testing/helpers/`, `ROADMAP.md`.
No production code.
