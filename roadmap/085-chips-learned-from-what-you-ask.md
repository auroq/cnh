# 85. Chips learned from what you ask: a profile chip is added when a question implies it, said in the open, removable

**Done 2026-11-09 (PR #101, issue #85). Filed 2026-11-09 by Aria, p1. Aria's call: the wording of what is learned and why.**

Depends on: 063, 066, 081

Numbered 085 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #85](https://github.com/auroq/cnh/issues/85); this file is the summary the roadmap keeps.

## What

When a question or a location implies a need (elevator or "can't do stairs" implies step-free; quiet room or "overwhelmed" implies quiet matters; sign language or captions implies captions or ASL; and so on), the matching chip is added to the profile, shown with a "from your question" marker and its own remove button, announced in the live region, and persisted like a chosen chip with a `learned` source flag; "Forget me" clears both.

## Why

Merge Forward: the profile should learn from what people ask without ever asking what they have, and say what it learned in the open.

## Acceptance

- Asking for the nearest elevator adds the step-free chip, announced; removing it works with keyboard only.
- A learned chip never overrides a chip the person removed in the same session.

## Touches

`pkg/ask`, `pkg/web`, `pkg/model`.
