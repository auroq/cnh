# cnh: roadmap

The plan for the one-day build and its current state.
The build is Option 1 from the kickoff plan, "Ask the venue": an attendee asks where the nearest step-free route, elevator, quiet room, all-gender restroom or captioned session is, and gets an answer that works with a screen reader, a keyboard and large text.
Option 2, "Find it in the talk", is a stretch goal and nothing in the rings depends on it.
The app is **Wide Door**, built by team **Curb Cut** (item [002](roadmap/002-the-app-needs-a-name.md), decided 2026-11-09); the repository stays `auroq/cnh`.

Every numbered item has its own file in [`roadmap/`](roadmap/); this file is the index and nothing else.
Decisions are in [`docs/adr/`](docs/adr/), and [`docs/README.md`](docs/README.md) says where everything else goes.

## Rings

Each ring ends with something deployable, so the project can stop at any ring and still be submitted.
Times are Mountain Time on 2026-11-09: coding starts at 11:00, code freezes at 17:00 and the submission is due at 17:30.

| ring | from | to | what must be true at the end | items |
| --- | --- | --- | --- | --- |
| 0 | 11:00 | 12:00 | the first commit after `pre-kickoff` is on `main`: README skeleton, AI_USAGE.md, `.mcp.json`, Gitar rules, and an Argo CD Application syncing a Radar-clean placeholder to the app namespace | 001, 003, 004, 005 |
| 1 | 12:00 | 13:00 | the app's Elasticsearch is up and the KubeCon schedule is indexed with room and level | 006, 007 |
| 2 | 13:00 | 14:00 | venue, room and access documents are seeded and the API answers "nearest X from room Y" and free-text questions | 008, 009 |
| 3 | 14:00 | 15:00 | the UI answers those questions with a screen reader, keyboard only, at 200% text and in high contrast | 010 |
| 4 | 15:00 | 16:00 | traces, metrics and logs reach Dash0; Radar shows no findings on the app; every change after Ring 0 went through a Gitar-reviewed PR | 011, 012, 004 |
| stretch | 16:00 | 17:00 | only if Ring 4 is done: transcript search inside talks, live mic as the default source; otherwise this hour is slack for Rings 0 to 4 | 014, 015, 016 |
| 5 | 17:00 | 17:30 | code frozen at 17:00; demo recorded, README architecture written, submission sent by 17:30 | 013 |

## Open

| item | what it wants | depends on | whose call |
| --- | --- | --- | --- |
| [013](roadmap/013-demo-and-submission.md) | the demo script, README architecture and the submission checklist | 010, 011, 012 | nobody |
| [014](roadmap/014-stretch-the-transcript-source.md) | **stretch:** where transcripts come from | 013 | Aria, the transcript source |
| [015](roadmap/015-stretch-live-mic-captions.md) | **stretch:** room-anchored large-text captions from a live mic | 014 | nobody |
| [016](roadmap/016-stretch-search-inside-talks.md) | **stretch:** timestamped transcript search, joined to the schedule | 007, 015 | nobody |
| [068](roadmap/068-venue-data-entrances-exits-first-aid-landmarks.md) | entrances by street side, exits and what is outside them, first aid, landmarks and aliases, each verified or not | 008 | Aria, what is verified |
| [072](roadmap/072-sessions-only-when-they-answer-the-question.md) | sessions listed only when the question is about them, with when and how far | 063, 075 | nobody |
| [079](roadmap/079-routes-not-just-destinations.md) | **after the event:** routes with obstacles, as steps in words and on a map | 068, 075, 081 | Aria, the next project |
| [083](roadmap/083-bring-your-schedule.md) | paste your Sched export or session titles, kept on your device, so answers know where you are going next | 081 | nobody |
