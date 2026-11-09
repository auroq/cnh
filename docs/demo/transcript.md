# Demo transcript, three minutes

Read this out loud while recording.
Lines in square brackets are what to have on screen, not what to say.
Target is about 420 words, which is three minutes at a normal talking pace.
Swap room names for wherever you are standing if you record at the venue.

## 0:00 The problem

[Phone, camera on you or on the Salt Palace signage.]

I'm Aria, team Curb Cut.
This is Wide Door.

Every conference tells you it's accessible.
Nobody tells you where the step-free door is from where you're standing right now.
The organiser's page lists the quiet rooms and the all-gender restrooms as text with no level and no route.
The venue's map doesn't say which restroom is single-occupancy.
And if you can't do stairs, "the elevator is by registration" isn't an answer at 4pm on Wednesday.

So I built the thing I'd want on my phone.

## 0:35 Ask the venue

[Phone, https://widedoor.cnh.auroq.com. Type "nearest all-gender restroom", location "255 BC".]

You say what you need and roughly where you are.
That's a room, a landmark, or the talk you just left.

[Read the answer.]

Same level, no stairs, and it tells me the organiser published this one.
That verified or unverified line is the whole point.
When we don't know, we say so.

[Type "I need to inject insulin".]

It understood a need, not a keyword.
Private medical space, nursing room first, first aid second, and the note tells you why.

## 1:10 The profile, in the open

[Tap Edit on the profile bar.]

This is what Wide Door knows about me.
Chips, not a diagnosis.
Step-free only.
Quiet spaces matter.
Stored on this phone until Thursday, sent with each question, never kept by the server.
If I ask for an elevator twice, it adds Step-free only and tells me it did, and I can remove it.
That came from a judge's suggestion this afternoon and it shipped two hours later.

## 1:35 Attendees close the loop

[Scroll to a place card, tap "Were you here?".]

Anyone standing at a door can say it works, or that the ramp is there but the door is locked.
A problem shows immediately as one attendee's report.
"Verified" takes two different people.
Rate limited, no account, random token.

## 1:55 How it's built

[Laptop, docs/architecture.md hexagon.]

Go, hexagonal.
The use case in `pkg/ask` declares a `Store` port and knows nothing about Elasticsearch.
`pkg/esindex` is the adapter.
The query is two retrievers fused with reciprocal rank, BM25 and ELSER semantic text, with a hard filter on the kind you asked for.
Needs and places live in YAML with a source on every fact.

[Kibana or the `places` index briefly.]

## 2:15 The pipeline and the sponsors

[GitHub PR 47, then PR 48.]

Every change went through a PR.
Gitar reviews against rules I wrote for accessibility, Kubernetes hardening and the test shape.
It refused a PR that deleted tests and approved the docs-only one with no human touching merge.
At 15:25 I made its block the merge gate.

[Dash0 dashboard, "Wide Door".]

Every question is one trace down to each Elasticsearch call.
The metric I care about is unverified answers served.
That's literally the list of doors a volunteer should go check.

[Radar, the cluster audit, then the rollout-abort page.]

Radar watched the cluster all day.
When a release shipped manifests ahead of the image, the canary stalled, Argo rolled it back, and Radar correlated it to the args change.
That was a real incident at noon, not staged.

## 2:45 Close

[Back to the phone.]

Wide Door works on a phone with a screen reader, a keyboard or big text.
It's what I'd hand a volunteer tomorrow.
Thanks.

## Recording notes

Two things to have open before you hit record: the live site on the phone and a laptop with four tabs in this order, `docs/architecture.md`, PR 47, the Dash0 dashboard, Radar.
If a page is slow on venue wifi, say the sentence anyway and keep moving.
If you run long, cut the pipeline section first, the profile section last.
