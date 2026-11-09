# 96. Attendees verify, dispute and annotate places: shown immediately, trusted after a second person agrees

**Open. Filed 2026-11-09 by Aria (issue #96); the form, the reports index, the trust rules and `widedoor reports` landed in PR #102, and the issue stays open for Aria to close. Aria's call: she defined the trust rules.**

Depends on: 063, 081 (only the cookie name)

Numbered 096 because issue numbers share a sequence with pull requests; the item number equals the issue number.
The full text, discussion and evidence are on [issue #96](https://github.com/auroq/cnh/issues/96); this file is the summary the roadmap keeps.

## What

Every place in an answer has a "Were you here?" form: works as described, a problem (with a note), or not as described (with a note).
Reports are documents in a `reports` index with the place id, the verdict, the note, the time and an anonymous per-browser token (`wd_id`).
A problem shows at once as "Reported 14:52 by one attendee: ...".
A place becomes "Verified by attendees (2)" when two different tokens agree it works.
An organiser-published place with a disputing report shows both.
Reports leave the display after six hours.
The summary says the attendee status in the same breath as verified or unverified.
Guard rails are server-side: rate limits per token and per IP, note length 10 to 280, no URLs, a honeypot, a blocked-word list, unknown place ids dropped.
`widedoor reports` lists and hides reports for volunteers.

## Why

Merge Forward: "the venue says it is accessible" is not trusted; what people want is verified fact from someone who was there.
Elasticsearch: reports are documents joined to places at query time.
Dash0: `widedoor.reports.received` and `widedoor.reports.rejected` say what attendees reported today.

## Acceptance

- A problem report on `step-free-200-south` makes the next answer include "Reported ... by one attendee: <note>".
- Two works-as-described reports from two tokens flip `elevators-hyatt` to "Verified by attendees (2)"; one does not.
- A fourth report from one token in ten minutes is rejected politely and counted.
- A note with a URL, a blocked word or five characters is rejected.
- The form works without JavaScript, with labelled controls, and the result is announced through the live region.
- `/api/ask` places carry `attendee_status`, `report_count`, `latest_note` and `latest_at`.

## Touches

`pkg/model`, `pkg/reports`, `pkg/esindex`, `pkg/ask`, `pkg/web`, `pkg/api/v1`, `pkg/cli`, `pkg/config`, `data/venue/blocked-words.txt`, `testing/`.
