# Transitional documents

A transitional document describes a state the project is passing through, not one it will stay in.
Examples: "demo data is still present", or "the transcript source is a live mic until the recordings are indexed".
It exists so that a temporary shortcut is written down where someone will find it, instead of being mistaken for the design.

Each one has:

- a **Status** line under its title: `Current since <date>`;
- a **Delete when** line: the condition that ends the state, specific enough to check;
- the state itself: what is true now, what is temporary about it, and what replaces it.

When the condition is met, the document is deleted in the same commit that meets it, and anything worth keeping moves into the README, an ADR or a roadmap item first.
A transitional document that outlives its condition is a bug.

## Current

| document | state | delete when |
| --- | --- | --- |
| [pre-kickoff-platform.md](pre-kickoff-platform.md) | `infra/` was built before the event and is described only here | the README's architecture section describes the platform |
