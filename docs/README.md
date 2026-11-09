# Where writing goes

| file | holds | does not hold |
| --- | --- | --- |
| [`../README.md`](../README.md) | what the app is, its architecture, how to run it, the tracks it enters | the plan or its status |
| [`../ROADMAP.md`](../ROADMAP.md) | the index: the rings and the open items | any item's detail |
| [`../roadmap/NNN-*.md`](../roadmap/) | one numbered item each: what, why, acceptance, what it touches | decisions that outlive the item |
| [`adr/`](adr/) | one decision each, with its context and consequences | how the work is going |
| [`transitions/`](transitions/) | a state the project is passing through, and when its doc is deleted | anything permanent |
| `../AI_USAGE.md` | how AI was used, per [ADR 0007](adr/0007-ai-disclosure.md) | anything else |

**An index stays an index.**
If detail is being added to `ROADMAP.md` or the README, it belongs in a file underneath them.

**Items and records are numbered and never renumbered.**
Roadmap items count from 001 and ADRs from 0001, and commits, PRs and issues cite them by number.

**An item's status is on line 3 of its file**: Open, Done or Adjudicated, with the date, and whose call it is.
`ROADMAP.md`'s open table is updated in the same commit.

**One sentence per line** in every Markdown file, so a diff shows the sentence that changed.
