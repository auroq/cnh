# Venue sources

Every fact in `data/venue/places.yaml` that came from the venue (Visit Salt Lake, operator of the Salt Palace) is listed here so a judge can audit it.
Verified means the venue states it in text on the page at the source URL.
Inferred means it was read off a drawing, a photo or a scene label, and nobody has checked it on the floor.
Attribute values follow the OpenStreetMap vocabulary: `wheelchair` is yes, limited, no or unknown; `entrance` is main, service or emergency; levels are numbers, with the venue's Level 1 as the lower level and Level 2 as the Mezzanine.
Extraction method: `tools/venue-extract/extract.py` fetches the text pages and the virtual tour script and prints candidate YAML, which was reviewed by hand before it went into `places.yaml`.

## Sources

| short name | URL | how it was read |
| --- | --- | --- |
| directory | https://www.visitsaltlake.com/salt-palace-convention-center/attend/venue-directory-map/ | text, by the script |
| parking | https://www.visitsaltlake.com/salt-palace-convention-center/attend/parking-information/ | text, by the script |
| tour | https://www.imapenterprises.com/saltpalace/virtualtour/index.htm | scene labels from its `script.js`, by the script; the panoramas themselves were not viewed |
| floor plan | https://www.visitsaltlake.com/salt-palace-convention-center/plan/interactive-floor-plan/ | not read; the CDN returns Access Denied to a non-browser fetch and the browser pass was not done before the hard stop |

## Facts

| fact | place id | source | status | OSM attribute |
| --- | --- | --- | --- | --- |
| The main street entrances are the East Entrance from West Temple and the South Entrance from 200 South | entrance-west-temple, entrance-200-south | directory | verified | entrance=main, wheelchair=unknown |
| There are also entrances on the North and West sides of the building | entrance-west, entrance-100-south | directory | verified for the sides; the 100 South street name is inferred | entrance=yes, wheelchair=unknown |
| Meeting rooms are numbered by floor; lower level rooms begin with 1, Mezzanine rooms begin with 2 | mezzanine-level-2 | directory | verified | level=1, level=2 |
| Only the northwest part of the building has a meeting room on the third floor | mezzanine-level-2, room-stack-155-255-355 | directory | verified | level=3 |
| Rooms 155, 255 and 355 are stacked on top of each other | room-stack-155-255-355 | directory | verified | level=1,2,3 |
| The easiest way to reach 155, 255 and 355 from the Main or South entrances is from the Mezzanine Level | room-stack-155-255-355, entrance-west-temple, entrance-200-south, step-free-200-south | directory | verified | level=2 |
| There are 10 exhibit halls, A through E running north to south and 1 through 5 running east to west | exhibit-halls | directory | verified | |
| The 200 South Lot and 300 West Lot are underground automated garages open 24 hours | parking-200-south-lot, parking-300-west-lot | parking | verified | amenity=parking, parking=underground |
| 200 South Lot entrances are off 200 South between West Temple and 200 West, and off 200 West just south of the dock entrance | parking-200-south-lot | parking | verified | |
| 300 West Lot entrances are off 300 West between 100 South and South Temple, and off 100 South past the Japanese Garden | parking-300-west-lot | parking | verified | |
| Accessible parking is available in both lots | parking-accessible, parking-200-south-lot, parking-300-west-lot | parking | verified; the count of spaces is not stated | wheelchair=yes |
| Bicycle racks are in the garages and at various places outside the facility | bike-racks | parking | verified | amenity=bicycle_parking |
| Parking is managed by ABM Parking Services, 385-468-2197 | parking-200-south-lot, parking-300-west-lot | parking | verified | |
| The 200 South Lot is nearest the South Entrance and the 300 West Lot is nearest the West entrances | parking-200-south-lot, parking-300-west-lot | parking, directory | inferred from the street names | |
| The tour has scenes named West Lobby, West Lobby Entrance, South Foyer, South Plaza Entrance, North Foyer, Entrance Foyer, Main Entrance, East Registration, Ballroom, 255 Foyer, 355 Foyer, Lower Mezzanine, Lower Mezzanine Entrance, Upper Mezzanine, Lower Concourse, Upper Concourse, Hall A B and C, Hall 1, Hall 4 and 5, Hall E, Room 255, Level 1, Level 2, Level 3 | west-lobby, south-foyer, north-foyer, east-registration, grand-ballroom | tour | inferred; a scene label is not a statement about position or access | |

## Restrooms

Neither text page mentions a restroom, so no restroom attribute was extracted from the venue.
The organiser's list of all-gender restrooms stays as it was, with its own source.
For every restroom the fields Aria asked for are not stated by the venue: all-gender or gendered, single-occupancy or multi-stall, wheelchair accessible, changing table, adult changing table.

## Connection facts for the route graph (#79)

These are the level changes and connections the sources support.
They are here even when they do not fit a place.

| from | to | by | source | status |
| --- | --- | --- | --- | --- |
| Main (East, West Temple) entrance | Mezzanine Level 2 | not stated; the venue says the Mezzanine is the easiest way to 155, 255 and 355 from this entrance | directory | verified that the connection exists; the means is not stated |
| South (200 South) entrance | Mezzanine Level 2 | not stated, as above | directory | verified that the connection exists; the means is not stated |
| Mezzanine Level 2 | Rooms 155 (Level 1), 255 (Level 2), 355 (Level 3) | stacked rooms, so a vertical connection exists at the stack | directory | verified that they are stacked; stairs, escalator or elevator not stated |
| Lower Mezzanine | Upper Mezzanine | not stated | tour | inferred from scene names only |
| Lower Concourse | Upper Concourse | not stated | tour | inferred from scene names only |
| Salt Palace | Hyatt Regency | hotel lobby, Level 1, by the entrance doors between the two buildings | organiser's inclusion page (already in places.yaml) | verified by the organiser, not by the venue pages |

## Not extracted before the hard stop

- The interactive floor plan, which is the only source likely to show elevators, escalators, first aid and emergency exits.
- The panoramas of the virtual tour, which might show whether the main doors have steps.
- Elevators and escalators by level, first aid, emergency exits with `leads_to`, info desks, and the Hyatt skybridge question.
- Restroom attributes per restroom.
