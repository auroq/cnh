---
title: "Architecture stays documented"
description: "A PR that adds a package or changes an import between layers updates docs/architecture.md"
type: check
when: "A PR adds a directory under pkg/ or cmd/, or changes an import of a github.com/auroq/cnh package in any .go file"
actions: "Check the PR also changes docs/architecture.md (and tools/archcheck.sh for a new package); if not, request changes naming the new package or the changed import and the section of docs/architecture.md that should describe it. Review the import against .gitar/review/architecture.md and quote the rule if it points outward."
---

# Architecture stays documented

`docs/architecture.md` is the page Aria presents; it must match the code.
When a PR adds a package, or changes which of this module's packages a file imports, it also updates `docs/architecture.md`, and `tools/archcheck.sh` when the package is new.
Review the import against `.gitar/review/architecture.md`: a use case importing an adapter, or an adapter importing a controller, is a blocker, quoted by file and line.
