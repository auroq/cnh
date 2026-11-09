# Go review

- Tests follow `.gitar/review/testing.md`: nested `t.Run` named "when ..., and ..., it should ...", one assertion in each leaf, tables for cases that differ only in data.
- Code lives in `cmd/<binary>/main.go` and `pkg/`; `main.go` only wires dependencies.
- Interfaces are declared by the package that uses them, not the one that implements them.
- Return errors with context (`fmt.Errorf("...: %w", err)`); never panic in library code.
- Every HTTP handler has a request-scoped context and a timeout on outbound calls.
- Spans, metrics and logs go through OpenTelemetry; no third-party logger without a bridge.
- No secrets in code, flags or test fixtures; configuration comes from the environment.
- No comments unless they record a decision the code cannot show; the code and the test names carry the meaning.
- Flags and scripts have `--help`; shell scripts start with `#!/usr/bin/env bash`.
- Code that exists only to satisfy a check, such as a health-check flag that inspects `/proc` for a batch job, is a finding, not a fix.
