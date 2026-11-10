# In my words

## Why I built this

When I read that the topic of this hackathon was related to improving conference experiences, I immediately thought accessibility.
Software and technology conferences gather together a uniquely diverse group with a single common interest,
but having diversity and a desire for inclusivity does not mean accessibility solves itself especially when you factor in intersectionality.
Recent technology focus on LLM and related automation tools has opened up a world where we can parse human speach in such a way to map it to live constantly changing real world data and situations.
*The moment or memory that made venue access the problem you picked over the other options in the research doc.*
*What "accessible" has meant to you at conferences before, and what it failed to mean.*
*Why "verified or unverified on every answer" mattered enough to be a rule.*

## What I decided today, and why

I decided to start this project by doing some prep work gathering information about real people and how conferences situations have or haven't met their needs.
Based on that data, I looked for intersectional needs where we could provide information to meet such needs without gathering unnecessary personal information.
As such, in the end I built "Wide Door" a conference "ask me anything" for accessibility needs.
The app initially started out stateless, but after a conversation with Kelsey Hightower,
I decided to use principles of ethical data collection to gather a minimal amount of information to create an anonymous profile that a user can use to get better results with the app.
Essentially, the data you provide alllows for more specific results; however, everything gathered is showed live to the user and can be immediately removed.

## How it's architected as I'd explain it at a whiteboard

I used hexagonal architecture, for this project as I usually tend to do for my projects.
`pkg/model` sits in the middle with the types the business actually means: Place, Session, Here, Answer, Preferences, Report.
`pkg/ask` is the use case.
It declares the ports it needs (`Store`, `IDLocator`, `SessionFinder`, `Status`) and it doesn't import Elasticsearch anywhere.
`pkg/esindex` is the adapter that implements those ports.
`pkg/sessionize` and `pkg/venue` are the other adapters, one reads the schedule feed and one reads my YAML.
`pkg/web` and `pkg/cli` are the controllers.
They wire things together and that's all they do.
`tools/archcheck.sh` enforces the import direction in CI, so the picture and the code can't drift apart.

A port is just an interface declared by the thing that needs it, not by the thing that provides it.
The one I'd swap first is `Store`.
That's the whole Elasticsearch surface, and the unit tests already run against a ten line fake of it.

Here's one question end to end.
You type "nearest elevator" and "255 BC".
`Locate` runs an ordered list of rules on "255 BC": door code id, exact room, alias, session title, fuzzy room, then semantic.
`KindOf` reads the question and says elevator.
If it hadn't matched a kind, `NeedOf` would try the needs vocabulary, then the semantic needs index with a score threshold I calibrated against the live cluster.
The store runs one RRF query with two retrievers, BM25 over name and notes and ELSER over the `semantic_text` field, with a hard filter on the kind.
`RankByNearness` sorts same level first, then same building, then everything else.
`Summarise` writes one sentence a screen reader hears first: the place, the level change, and whether anyone verified it.
If you had a profile, the chips ran right after `KindOf` and you'd see which ones changed the answer.

Where a type lives is a three way question.
If it's what the business means, it's in `pkg/model`.
If it's how one adapter stores or reads it, it's private to that adapter.
If it's what clients were promised, it's in `pkg/api/v1`, frozen, and there's a byte for byte recorded response test that fails if anyone touches it by accident.

## How I used the sponsors' tools

Gitar reviewed every PR after the first commit.
I wrote the rules it reviews against: `.gitar/review/a11y.md`, `k8s.md`, `testing.md`, `architecture.md`.
It refused PR 47, which deleted tests with no linked issue, and approved the docs only PRs with no human touching merge.
The honest part is the middle of the day.
At 13:47 I made its approval advisory because a single author can't approve their own PR and Gitar's approval step couldn't read PR bodies on this plan.
Nine PRs merged with findings open in that window.
At 15:25 I turned on its merge blocking for important findings and made that the gate.
It blocked PR 101 ten minutes later for two real problems and lifted the block when they were fixed.
ADR 0010 and 0013 tell that story in order.

Dash0 gets traces, metrics and logs over OTLP from one adapter package behind one config value.
The trace I open first is a `GET /api/ask`.
Under it is `ask.Ask` with the question length, the resolved venue and level, and whether the top answer was verified, then the Elasticsearch calls as children.
The metric I watch is `widedoor_ask_unverified_served_total`.
It's the number of times we told someone where a door is when nobody had checked, which is literally the list a volunteer should walk.
The dashboard is a Perses JSON file in the repo, pushed with `make dashboard`, so it's reviewed like code.

Radar audited the cluster all day.
The `wide-door` namespace is clean except two things I accepted on purpose and wrote down: Elasticsearch and Kibana run without a read only root filesystem because they write under their install directory, and ECK generates two client cert secrets nothing mounts because TLS stops at the gateway.
At noon a release shipped manifests before the image that understood them.
The canary stalled, Argo Rollouts aborted it, the old pods kept serving, and Radar flagged the stall and correlated it to the args change.
That was a real incident.
It's written up in `docs/demo/2026-11-09-rollout-abort.md`.

Elasticsearch is the only store.
I use two retrievers because they fail differently.
BM25 is exact and literal, so "255" finds room 255.
ELSER is semantic, so "I need to inject insulin" finds the nursing room without the word nursing in it.
Reciprocal rank fusion lets both vote.
`semantic_text` gave me the semantic side with no embedding pipeline of my own, and the trial licence got me ELSER in an afternoon.
Needs are indexed the same way, and the multilingual E5 endpoint is already on the cluster for the next step.

## What I did with my hands

I designed and wrote drafts of the issues and let claude ask me questions to flush out more details before putting it into github issues..
Every feature after the first ring started as an issue in my words with a what, a why and an acceptance list, and the agents built to it.
The ones that mattered most were the ones I argued for: kind before nearness after the acceptance table found the elevator bug, needs instead of keywords, location in your own words, QR codes on doors, the profile chips, learned chips, attendee verification with the two person rule, and the restroom attributes I didn't get to.

I talked to a judge and changed the product because of it.
Kelsey Hightower told me the next level was an anonymous profile done ethically, shown on the screen.
Two hours later it shipped.

I made the data rules.
Verified means a publisher said it in text, the organiser or the venue.
Anything read off a drawing is unverified and says so.
Nothing in the app implies a restroom is accessible or single occupancy unless a source says so.
I'd rather tell a wheelchair user "we don't know" than guess.

I reviewed and I overruled.
I turned the Gitar gate off when it couldn't work and back on when it could.
I killed walk time estimates after the research said disabled travellers want route facts, not minutes.
I renamed a chip that named a person instead of a need.
I pulled routes out of production for today because a half verified route is worse than none.
I moved the schedule import to post event and then put it back when the only blocker was a merge conflict.

## What I'd do next

Routes with obstacles.
The graph and the constrained search exist in PR 125.
What's missing is Level 3 to 5, exits, and a map with the same steps as its text alternative.

Restrooms described by what they offer.
All gender, single occupancy, wheelchair, changing table, adult changing, each one stated or "not stated".

A few words in any language.
The vocabulary in Japanese, Korean, Spanish and Chinese, and the multilingual E5 field as a third retriever.

Off, Prefer, Must as a real control.
The behaviour shipped today.
The radio groups on the profile page didn't.

Attendees adding places.
Behind a quality gate, held unverified until a second person confirms, with a volunteer review queue.

The schedule by account.
Sched or Sessionize import with consent, and the same chips deciding what the app may suggest.
