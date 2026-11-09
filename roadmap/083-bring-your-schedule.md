# 83. Bring your schedule: paste your Sched export or session titles, kept on your device

**Open. Filed 2026-11-09 by Aria, p2, ready for a worker. Nobody's call once item 081 lands.**

Depends on: 081

Numbered 083 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #83](https://github.com/auroq/cnh/issues/83); this file is the summary the roadmap keeps.

## What

A `/schedule` page with one textarea for a pasted Sched iCal export or a list of session titles; titles are matched to the sessions index (exact first, then best match above a threshold), the matched ids are kept in the browser only, and the page shows "Your next session" and "You just left", defaults the location to the room you just left, and offers the nearest quiet room when three sessions run back to back.
`/api/ask` accepts `sched` and returns `schedule.next`, `schedule.just_left` and `schedule.assumed_location`.

## Why

The judge's "next level" without an account: answers shaped by where the person is going, with the same ethics as the profile.
Elasticsearch: title matching from messy pasted text is a retrieval problem.

## Acceptance

- A pasted list of three titles matches three sessions and shows the next one with its room.
- Nothing about the schedule is stored server-side.

## Touches

`pkg/web`, `pkg/ask`, `pkg/esindex/store.go`, `pkg/api/v1`.
