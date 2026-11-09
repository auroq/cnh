# 32. Configuration through cobra and viper, one binary, every layer replaceable

**Open. Filed 2026-11-09, during Ring 3, by Aria. Aria's call: the shape mirrors her cellar and botropolis repos.**

Depends on: 025, 031

Numbered 032 because issue numbers share a sequence with pull requests; the item number equals the issue number.

## What

- One binary, `widedoor`, with cobra subcommands `serve` and `ingest`; `cmd/widedoor/main.go` is `os.Exit(app.Main())`.
- `pkg/config`: a `Key*` constant for every setting, `Defaults` in one map, `New()` (defaults, `WIDEDOOR_` environment with dots as underscores, config file search path), `BindFlags` (a flag per non-secret setting), `Load()` (flag beats env beats file beats default), and typed interfaces per area: `Server`, `Elasticsearch`, `Ingest`, `Telemetry`. The Elasticsearch password comes only from the environment or the file, never a flag, and is never logged.
- `pkg/cli`: the root command with a `Run:` group, `serve` and `ingest` that only load config, construct adapters and hand them to use cases; `pkg/app.Main()` turns the result into an exit code (2 when the store rejected documents).
- `pkg/ingest`: the ingest use case with its ports `Source` (the schedule), `Places` (the curated file) and `Indexer` (the store's write side); `pkg/schedule` holds the `Session` model so the use case imports no adapter.
- Deploy: non-secret settings live in ConfigMap `widedoor-config` mounted as `/etc/widedoor/config.yaml` for both the Rollout (`serve`) and the CronJob (`ingest`); the password stays a `secretKeyRef`. One image, `ghcr.io/auroq/cnh/widedoor`.

## Why

Every layer takes the interface it needs, so a test hands it a fake and a judge can see where any setting comes from.
One binary means one image, one bump per release and one `--help` that lists every setting with its environment name.
The `Telemetry` interface is the hook item 011 plugs into.

## Acceptance

- `widedoor --help` lists both commands and every setting with its default and environment name.
- `go test ./pkg/config` proves flag beats env beats file beats default, and that the password is not a flag.
- `go test ./pkg/ingest` runs the use case against fakes for all three ports.
- `tools/archcheck.sh` passes with the new packages placed.
- The CronJob and the Rollout run the same image with `ingest` and `serve` as arguments.

## Touches

`cmd/widedoor`, `pkg/app`, `pkg/cli`, `pkg/config`, `pkg/ingest`, `pkg/schedule`, `pkg/esindex`, `pkg/sessionize`, `pkg/venue`, `deploy/app`, `Dockerfile`, `.github/workflows/images.yml`, `tools/archcheck.sh`.
