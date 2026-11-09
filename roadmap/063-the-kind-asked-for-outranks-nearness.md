# 63. The kind the attendee asked for outranks nearness

**Done 2026-11-09 (PR #71, issue #63). Filed 2026-11-09 by Aria. Nobody's call: the acceptance table in item 054 showed the bug and the fix is the use case asking its port for the kind it inferred.**

Depends on: 025, 054

Numbered 063 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

When the question names a kind of place (elevator, quiet room, prayer room, all-gender restroom, step-free entrance, help desk, coat check, nursing room), the top answer is a place of that kind.
Nearness orders places within the kind, never across kinds.
Before this, `ask.RankByNearness` moved any same-level or same-building place ahead of the kind asked for, so with a room given the answer was often the wrong thing nearby.

## Why

Merge Forward: a wheelchair user asking for the nearest elevator from 255 BC was told where the coat check is.
That is the one answer this app must never get wrong, and it is judge-visible in the demo.
Found by the acceptance table work in item 054; its issue comment lists the live answers.

## Acceptance

- `ask.KindOf` infers a kind from the question using the quick-pick vocabulary and obvious synonyms: lift and elevator, toilet, restroom and bathroom, quiet and sensory, pray and prayer, nursing and lactation, help, info and registration, coat and bag, step-free and accessible entrance.
- When a kind is inferred, the use case asks its `Store` for places of that kind only (`FindPlaces` takes the kind; `esindex.Store` filters on the `kind` keyword), which also fixes the "other building" case because the right place is always in the result.
- `Summarise` keeps saying same level, same building or other building, and verified or unverified.
- Unit tests in `pkg/ask` for the inference table and for the kind reaching the store; a unit test in `pkg/esindex` for the filter in both retrievers; an integration test under `testing/integration/ask` that "nearest elevator" from 255 BC answers the elevator and not the restroom on the same level.
- The rows reported on issue #54 are in `testing/acceptance/cases.go` and pass against the live app after deploy; "prayer room" from 255 BC exercises the "It is in the Hyatt Regency, you are in the Salt Palace" wording.
- Questions that name no kind, such as "sign language", behave as before.

## Touches

`pkg/ask` (`kind.go`, the `Store` port), `pkg/esindex/store.go`, `testing/integration/ask`, `testing/acceptance/cases.go`, `docs/architecture.md`.
