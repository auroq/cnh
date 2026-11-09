# 50. Split config

**Open. Filed 2026-11-09 by Aria, without a roadmap item at the time; filed here in the roadmap sync. Aria's call: whether to do it at all, p3.**

Depends on: 032

Numbered 050 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #50](https://github.com/auroq/cnh/issues/50); this file is the summary the roadmap keeps.

## What

`pkg/config/config.go` holds every key, default, interface and accessor in one file.
Split it by area (server, elasticsearch, ingest, telemetry) for readability, keeping `Key*` constants and `Defaults` discoverable from one place.

## Why

Readability for the person explaining the configuration; no behaviour change.

## Acceptance

- `go test ./pkg/config` passes unchanged.
- `widedoor --help` prints the same flags.

## Touches

`pkg/config/`.
