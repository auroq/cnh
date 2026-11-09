---
title: "Tests ship with code"
description: "Ask for tests when Go code changes without test changes"
type: check
when: "A PR changes a non-test .go file but no _test.go file"
actions: "Post one PR comment naming the changed functions and asking for a test in the shape of .gitar/review/testing.md that covers the new or changed behavior"
---

# Tests ship with code

When a pull request changes Go source without touching any `_test.go` file,
post one comment naming the changed functions and asking for a test that covers the new or changed behavior,
nested as "when ..., and ..., it should ..." with one assertion per leaf, as `.gitar/review/testing.md` describes.
A bug fix needs a case that fails without the fix.
