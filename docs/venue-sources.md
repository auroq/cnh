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
| floor plan | https://www.visitsaltlake.com/salt-palace-convention-center/plan/interactive-floor-plan/ | the CDN returns Access Denied to a non-browser fetch, so Levels 1 and 2 were read in Claude in Chrome with the icon layer on; Level 3 was not opened |

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

## Floor plan, Levels 1 and 2

| fact | place id | status | OSM attribute |
| --- | --- | --- | --- |
| The north and south ballroom foyers provide reception and registration space; the South Plaza is an outdoor area | south-foyer, north-foyer | verified, stated in the Level 1 caption | |
| Level 2 has a large sunlit lobby that leads to the Business Center, the Visit Salt Lake Visitors' Center, the gift shop and meeting rooms | visitor-center-level-2, mezzanine-level-2 | verified, stated in the Level 2 caption | level=2 |
| First aid icons at the North Foyer and at the Lower Mezzanine by room 155 | first-aid-north-foyer, first-aid-lower-mezzanine | inferred from icons | emergency=first_aid |
| Security icon between room 151 and Hall A to E | security-151 | inferred from icons | |
| Elevator icons on Level 1 at South Foyer, North Foyer, Lower Mezzanine, the Hall 1 to 5 concourse and West Lobby; on Level 2 at East Registration, Upper Concourse by 250, Upper Mezzanine by 255 and 257, and West Lobby | elevators-level-1, elevators-level-2 | inferred from icons; levels served not shown | highway=elevator |
| Escalator icons on both levels at South Foyer, North Foyer, Lower and Upper Mezzanine, and West Lobby | escalators-level-1 | inferred from icons | conveying=yes |
| Entrance icons on Level 1 at South Plaza (two), the north end of Hall A to E, the Lower Mezzanine east side (two) and West Lobby (two); on Level 2 at East Registration (two) | entrance-200-south, entrance-west-temple, entrance-west, entrance-100-south | inferred from icons; step-free not shown | entrance=yes, wheelchair=unknown |
| Parking icons at the north end of Hall A to E and at the West Lobby | parking-300-west-lot | inferred from icons | |
| Restroom icons along both hall concourses, in both foyers, at the Lower Mezzanine and West Lobby, at East Registration, and by 255 and 257 | restrooms-level-1, restrooms-level-2 | inferred from icons; no attribute stated | |

## Restrooms

Neither text page mentions a restroom, so no restroom attribute was extracted from the venue.
The floor plan shows restroom icons only, so `restrooms-level-1` and `restrooms-level-2` carry every attribute as not stated.
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
