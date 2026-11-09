# Go review

- Tests are table-driven subtests named "when [scenario], and [context], it should [expectation]", one assertion each.
- Code lives in `cmd/<binary>/main.go` and `pkg/`; `main.go` only wires dependencies.
- Interfaces are declared by the package that uses them, not the one that implements them.
- Return errors with context (`fmt.Errorf("...: %w", err)`); never panic in library code.
- Every HTTP handler has a request-scoped context and a timeout on outbound calls.
- Spans, metrics and logs go through OpenTelemetry; no third-party logger without a bridge.
- No secrets in code, flags or test fixtures; configuration comes from the environment.
