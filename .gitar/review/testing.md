# Testing standard

These rules apply to every test in the repository, in any language, whether a person or an auto-apply wrote it.
A test that breaks one of them is a review finding, quote the test by name.

## Naming and shape

- Tests read as behaviour: "when [scenario], and [context], it should [expectation]".
- In Go that is nested `t.Run` blocks: the outer `t.Run("when ...")` holds the setup for the scenario, an inner `t.Run("and ...")` narrows it, and the innermost `t.Run("it should ...")` holds exactly one assertion.
- A flat test function such as `TestFooSkipsBar` with several `if` checks in a row is the wrong shape; ask for it to be split into the hierarchy above.
- One assertion per innermost test. If two things must be true, that is two `it should` blocks sharing the parent's setup.
- Setup lives in the parent scope, not repeated in every leaf.
- Cases that differ only in data are a table iterated over, with the case name in the `t.Run` label; nest tables when it avoids duplication (for every input size, for every file type).
- Test names say what the behaviour is, not what the function is called.

## Independence

- Every test is idempotent, safe to run in parallel with any other, and depends on no external state: no network, no real Elasticsearch, no clock, no filesystem outside a temp dir.
- Every test cleans up after itself (`t.TempDir`, `t.Cleanup`, `httptest.NewServer` closed).
- Unit tests test only the unit; dependencies are faked at the interface the unit itself declares.
- Unit tests sit next to the code they test, `foo_test.go` beside `foo.go`.
- Tests that cross units or talk to a real system go under `testing/integration/`, end-to-end tests under `testing/acceptance/`, shared setup under `testing/helpers/`, and the directory tree under each mirrors the code it tests.

## What a PR must carry

- Code and its tests arrive in the same PR; a fix without a test for the behaviour it fixes is incomplete.
- A bug fix carries a test case that fails without the fix, named for the bug as a behaviour, not for the fix.
- No test is skipped, commented out or weakened to make CI pass.
- No comments in tests or code unless they explain a decision that cannot be read from the code.
