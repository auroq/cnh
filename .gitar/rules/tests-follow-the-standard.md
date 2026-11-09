---
title: "Tests follow the standard"
description: "Review every added or changed test against the testing standard"
type: check
when: "A PR adds or changes a _test.go file, or any file under testing/"
actions: "Review against .gitar/review/testing.md; request changes quoting each test that is flat-named, holds more than one assertion in a leaf, repeats setup a parent could hold, depends on external state or does not clean up"
---

# Tests follow the standard

Every test is written as "when [scenario], and [context], it should [expectation]", nested, with one assertion in each leaf and tables for cases that differ only in data.
Check every added or changed test against `.gitar/review/testing.md`.
Quote the test by name, say which rule it breaks, and show the nested shape it should take.
A test that talks to the network, a real Elasticsearch or the clock in a unit test is a blocker.
