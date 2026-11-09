# 57. Types live by kind: domain model, adapter shapes, and a versioned API contract

**Open. Filed 2026-11-09, during Ring 4, by Aria. Nobody's call: it finishes what items 025 and 031 started.**

Depends on: 031, 054

Numbered 057 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

- `pkg/model` holds `Place`, `Room`, `Session`, `Here` and `Answer` and imports nothing in the module; every port speaks model types; `venue.Place` and `ask.Place` collapse into `model.Place`, `schedule.Session` and `ask.Session` into `model.Session`.
- Adapter shapes stay private to their adapter: the YAML file shape in `pkg/venue`, the Elasticsearch documents in `pkg/esindex`, the markdown rows in `pkg/sessionize`; each maps at its boundary.
- `pkg/api/v1` is the public JSON contract of `/api/ask`, imports nothing, field names frozen; `pkg/web` maps the model onto it, and only `pkg/web` imports it.
- `tools/archcheck.sh` enforces the three homes; `docs/architecture.md` and `.gitar/review/architecture.md` say where a type lives; ADR 0011 records it.

## Why

Rule 6: Aria explains where a type lives with one of three sentences, "it is what the business means", "it is how this adapter stores it", "it is what clients were promised".
It also removes the duplicated `Place`, and the acceptance tests of item 054 become the proof that the contract did not move.

## Acceptance

- `tools/archcheck.sh` passes with the new rules and fails if anything but `pkg/web` imports `pkg/api/v1` or `pkg/model` imports anything in the module.
- The JSON of `/api/ask` is byte-for-byte identical before and after, checked against a recorded response in `pkg/web`.
- Existing tests pass unchanged except for import paths.

## Touches

`pkg/model` (new), `pkg/api/v1` (new), `pkg/ask`, `pkg/venue`, `pkg/schedule`, `pkg/esindex`, `pkg/web`, `tools/archcheck.sh`, `docs/architecture.md`, `.gitar/review/architecture.md`, `docs/adr/0011-types-live-by-kind.md`.
