# 0012. How we describe distance: route facts, never a single walk time

Status: Accepted, 2026-11-09. Nobody's call: it follows published practice, roadmap item 075 (issue #75).

## Context

Wide Door answers "where is the nearest X from here" for attendees with limited mobility, hearing or vision.
Item 072 proposed an estimated walk time per session ("about 3 minutes on the same level, plus 3 for a level change, plus 8 for the other building").
A single time assumes a pace and a body.
A wheelchair user, someone with chronic fatigue, a blind attendee navigating by cane and a parent with a stroller each take different routes and times, and a wrong estimate is worse than none, because the people this app is for are the ones who cannot absorb the error.
Aria asked for thirty minutes of research into what accessibility and wayfinding practice actually publishes before any wording was written.

## What the practice says

Conference and event guidance asks for route facts, not minutes.

- The SIGACCESS Accessible Conference Guide asks organisers to describe walking distance, terrain, steps and lifts, and rest points between rooms; it does not ask for a time.
- The W3C WAI guide to making events accessible puts accessible routes first and says only that moving between rooms "may take more time", with no figure.

Venue access guides written by and for disabled people publish measurements and features.

- Euan's Guide tells venues to list whether everything is on one level, which routes are step-free, and where the lifts are.
- AccessAble's Detailed Access Guides give the distance in metres from a landmark ("21m from the main entrance"), step heights, surfaces and kerbs, and no time.

Transit step-free guidance describes the gap, the step and the lift, and treats time as a coarse band at most.

- Transport for London's step-free Tube guide (October 2026 edition) gives the step and gap between platform and train in millimetres, which stations have lifts, and street-level interchanges in metres; minutes appear only as a symbol for "under 10 minutes" or for long transfers, with no range and no pace.
- The UK Department for Transport's Inclusive Mobility guide (2021, page 25) gives recommended distances without a rest by impairment, from 50 m for people using a stick or cane to 150 m for wheelchair users and people with impaired vision, and says there is much variation between individuals.
- A British Columbia report on hidden mobility disability (2019) found no clear relationship between the distance people could cover and the time they said it took, that 39 percent could walk two minutes or less, and recommended concrete distance cues such as "a school bus length" instead.

Map data and routing engines publish facts and let the person set their own limits.

- OpenStreetMap's `wheelchair`, `step_count`, `incline`, `surface`, `kerb` and `width` keys and the Simple Indoor Tagging scheme (`level`, `highway=elevator`, `highway=steps`) record what is there.
- The OpenRouteService wheelchair profile takes maximum incline, maximum kerb height, minimum width and surface type as parameters, and has no speed or pace setting.

One counterpoint, recorded so nobody thinks it was missed.

- The W3C Cognitive Accessibility Guidance for Wayfinding (draft note, February 2026) recommends telling the user the time to the destination up front, notes that metres are hard for some people with cognitive disabilities, and asks that units be customisable and route differences such as "more steps" be shown.
  It argues for a time the person can interpret, not for a default pace applied to everyone, and its request for customisable units is the same request as the pace setting below.

## Decision

1. Distance is described as route facts, in this order: the tier in words (same level; one level up or down, by elevator or stairs; the other building), then the facts we know about the way there (connected at the hotel lobby, step-free, through a crowd zone, a rest point on the way), then whether anyone verified it.
2. A distance in metres, with a familiar comparison, is given when we know it.
   Today we know none, so none is shown.
3. No single walk time is shown, ever, by default.
   If a time is shown in a later item, it is a coarse band ("under 5 minutes", "5 to 15 minutes") tied to a pace the attendee chose from at least three options and remembered in their browser, and the band is labelled with that choice.
4. "We do not know" is said when we do not know.
   A place on an unknown level in the attendee's building is "we do not know which level it is on", not a guess.
5. The organiser does not publish elevator positions, so the app does not either; the route fact is that registration staff will point to the nearest one.

## Consequences

- `pkg/ask.Summarise` says "Same building, one level up, by elevator or stairs" and counts the levels; the walk-time estimate in item 072 is dropped before it ships, by Aria's note on that issue.
- `model.Place` gains `Route`, a list of plain sentences curated in `data/venue/places.yaml`, indexed in Elasticsearch, shown under each place and read into the summary for the first one; `pkg/api/v1` gains the optional `route` field, which is additive and leaves every recorded response byte for byte the same.
- The only route facts in the data today come from the organiser's inclusion page: the Hyatt Regency is connected to the Salt Palace, at the hotel lobby on Level 1, and nobody has published where the elevators are.
  Anything else waits for someone to walk it.
- A later item can add metres and the "my pace" setting without changing the sentence shape.

## Sources

- SIGACCESS Accessible Conference Guide: https://www.sigaccess.org/welcome-to-sigaccess/resources/accessible-conference-guide/
- W3C WAI, How to Make Your Presentations and Meetings Accessible to All: https://www.w3.org/WAI/teach-advocate/accessible-presentations/
- Euan's Guide, tips for listing a venue: https://www.euansguide.com/tips-for-listing-venue
- AccessAble Detailed Access Guides: https://www.accessable.co.uk/
- Transport for London, Step-free Tube guide: https://content.tfl.gov.uk/step-free-tube-guide-map.pdf
- UK Department for Transport, Inclusive Mobility (2021): https://www.gov.uk/government/publications/inclusive-mobility-making-transport-accessible-for-passengers-and-pedestrians
- Government of British Columbia, Hidden Mobility Disabilities report (2019): https://www2.gov.bc.ca/assets/gov/government/about-the-bc-government/accessibility/hidden_mobility_disability_report_2019.pdf
- OpenStreetMap wiki, Key:wheelchair and Simple Indoor Tagging: https://wiki.openstreetmap.org/wiki/Key:wheelchair and https://wiki.openstreetmap.org/wiki/Simple_Indoor_Tagging
- OpenRouteService wheelchair profile: https://giscience.github.io/openrouteservice/api-reference/endpoints/directions/routing-options
- W3C, Cognitive Accessibility Guidance for Wayfinding (draft note): https://www.w3.org/TR/coga-wayfinding/
- KubeCon NA 2026 inclusion and accessibility page (the only venue route facts): https://events.linuxfoundation.org/kubecon-cloudnativecon-north-america/attend/inclusion-accessibility/
