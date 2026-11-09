# 0002. Build "Ask the venue"; searching inside talks is a stretch

Status: Accepted, 2026-11-09.

## Context

The app built from Option 1 is named Wide Door (roadmap item 002).

The kickoff plan ranked three options from the accessibility research.
Option 1, "Ask the venue": attendees ask where the nearest step-free route, elevator, quiet room, all-gender restroom or captioned session is.
Option 2, "Find it in the talk": timestamped transcript search across talks.
Option 3, "Change watch": notify attendees when a session's time or room changes.
The research found the same failure across every group it covered: the information exists but nobody can look it up from where they stand.
The organiser publishes room lists as static text, and the venue's OpenStreetMap record has no accessibility tags.
There is one developer and about six hours of build time.

## Decision

Build Option 1.
Option 2 is a stretch goal for the 16:00 to 17:00 hour, only if Ring 4 is done, and the plan assumes it never starts.
Option 3 is not built.

## Consequences

- Every sponsor track can be entered with Option 1 alone: Elasticsearch for search, Dash0 and Radar for operations, Gitar for the PR flow, Merge Forward for the access focus.
- The venue data is partly unverified, and the app has to say so in every answer that uses it.
- Hearing access is weaker without Option 2; the app covers it only by pointing to captioned sessions.
