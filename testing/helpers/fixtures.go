package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/pkg/ingest"
	"github.com/auroq/cnh/pkg/schedule"
	"github.com/auroq/cnh/pkg/venue"
)

const sessionizeURL = "https://sessionize.com/api/v2/uo948i5p/view/markdown"

const inclusionURL = "https://events.linuxfoundation.org/kubecon-cloudnativecon-north-america/attend/inclusion-accessibility/"

var MST = time.FixedZone("MST", -7*3600)

var Places = []venue.Place{
	{ID: "room-salt-palace-level-2-255-bc", Kind: "room", Name: "255 BC", Venue: "Salt Palace", Level: "Level 2", Near: "Salt Palace | Level 2 | 255 BC", Verified: true, Source: sessionizeURL},
	{ID: "room-salt-palace-level-1-grand-ballroom-j", Kind: "room", Name: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1", Near: "Salt Palace | Level 1 | Grand Ballroom J", Verified: true, Source: sessionizeURL},
	{ID: "agr-sp-l2-255", Kind: "all-gender-restroom", Name: "All-gender restroom outside 255", Venue: "Salt Palace", Level: "Level 2", Near: "Outside room 255", For: []string{"restroom", "trans", "non-binary"}, Verified: true, Source: inclusionURL},
	{ID: "agr-sp-l1-showcase", Kind: "all-gender-restroom", Name: "All-gender restroom, Solutions Showcase", Venue: "Salt Palace", Level: "Level 1", Near: "Solutions Showcase, Halls A-C and 1-5, Amber Hive Zone", For: []string{"restroom", "trans", "non-binary"}, Verified: true, Source: inclusionURL},
	{ID: "agr-hy-l1-lobby", Kind: "all-gender-restroom", Name: "All-gender restroom, Hyatt lobby", Venue: "Hyatt Regency", Level: "Level 1", Near: "Lobby, by the Salt Palace entrance doors", For: []string{"restroom", "trans", "non-binary"}, Verified: true, Source: inclusionURL},
	{ID: "quiet-room-260a", Kind: "quiet-room", Name: "All-attendee quiet room", Venue: "Salt Palace", Level: "Level 2", Near: "Room 260A, Level 2 meeting rooms", Notes: "Open to every attendee. No talking, no interacting, no working inside.", For: []string{"sensory", "rest", "neurodivergent"}, Verified: true, Source: inclusionURL},
	{ID: "community-hub-grand-j", Kind: "community-hub", Name: "Community Hub and Merge Forward gatherings", Venue: "Salt Palace", Level: "Level 1", Near: "Grand Ballroom J", Notes: "Peer group gatherings, Sign Language Crash Course and Merge Forward sessions meet here on 10 to 12 November.", For: []string{"community", "deaf", "queer", "bipoc", "women"}, Verified: true, Source: sessionizeURL},
	{ID: "registration-200-south", Kind: "info-desk", Name: "Registration and badge pick-up", Venue: "Salt Palace", Near: "200 South entrance", Notes: "Staff here know the elevator locations, the nursing room and the first-aid points.", For: []string{"help", "mobility"}, Verified: true, Source: sessionizeURL},
	{ID: "nursing-room", Kind: "nursing-room", Name: "Private nursing room", Venue: "Salt Palace", Near: "Ask at registration or any information desk for the location", Notes: "A private nursing room is available on every event day.", For: []string{"parents", "nursing"}, Verified: true, Source: inclusionURL},
	{ID: "elevators-salt-palace", Kind: "elevator", Name: "Elevators between Salt Palace levels", Venue: "Salt Palace", Near: "The organiser does not publish elevator positions. Staff at registration and the information desks will point you to the nearest one.", Notes: "Seeded as unverified.", For: []string{"mobility"}, Verified: false, Source: "floor plan, unverified"},
}

var Needs = []venue.Need{
	{ID: "mobility-accessible", Name: "a step-free way around", Phrases: []string{"wheelchair", "can't do stairs", "step free"}, Kinds: []string{"step-free-entrance", "elevator"}},
	{ID: "private-space", Name: "somewhere private", Phrases: []string{"fix my makeup", "change clothes", "somewhere private"}, Kinds: []string{"all-gender-restroom", "quiet-room"}, Note: "Single-stall where known."},
	{ID: "medical-private", Name: "a private place for medication", Phrases: []string{"insulin", "injection", "inject", "medication"}, Kinds: []string{"nursing-room", "first-aid", "all-gender-restroom"}, Note: "A private room; first aid staff can help."},
	{ID: "rest", Name: "somewhere to rest", Phrases: []string{"sit down", "rest", "overwhelmed"}, Kinds: []string{"quiet-room"}},
}

var Sessions = []schedule.Session{
	{
		ID: "1292207", Title: "Coat + Bag Check", URL: sessionizeURL + "/session/1292207",
		Start: time.Date(2026, 11, 9, 7, 0, 0, 0, MST), End: time.Date(2026, 11, 9, 19, 0, 0, 0, MST),
		Location: "Salt Palace | Level 2 | East Registration", Venue: "Salt Palace", Level: "Level 2", Room: "East Registration",
		Description: "Leave your coat and bags at East Registration and pick them up before 7 pm.",
	},
	{
		ID: "1400001", Title: "Sign Language Crash Course", URL: sessionizeURL + "/session/1400001",
		Start: time.Date(2026, 11, 11, 13, 30, 0, 0, MST), End: time.Date(2026, 11, 11, 14, 15, 0, 0, MST),
		Location: "Salt Palace | Level 1 | Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1", Room: "Grand Ballroom J",
		Description: "Learn the ASL alphabet and a handful of signs to greet Deaf and hard of hearing attendees at the Community Hub.",
	},
	{
		ID: "1400002", Title: "Keynote: Welcome and Opening Remarks", URL: sessionizeURL + "/session/1400002",
		Speakers: []string{"Someone Else"},
		Start:    time.Date(2026, 11, 10, 9, 0, 0, 0, MST), End: time.Date(2026, 11, 10, 9, 10, 0, 0, MST),
		Location: "Salt Palace | Level 2 | 255 BC", Venue: "Salt Palace", Level: "Level 2", Room: "255 BC",
		Description: "The co-chairs open KubeCon + CloudNativeCon North America 2026 and set out the week.",
	},
}

func LoadPlaces(t *testing.T, es *esindex.Client, index string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	x := &esindex.Indexer{Client: es, PlacesIndex: index}
	check(t, index, len(Places))(x.IndexPlaces(ctx, Places))
}

func LoadNeeds(t *testing.T, es *esindex.Client, index string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	x := &esindex.Indexer{Client: es, NeedsIndex: index}
	check(t, index, len(Needs))(x.IndexNeeds(ctx, Needs))
}

func LoadSessions(t *testing.T, es *esindex.Client, index string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	x := &esindex.Indexer{Client: es, SessionsIndex: index}
	check(t, index, len(Sessions))(x.IndexSessions(ctx, Sessions))
}

func check(t *testing.T, index string, want int) func(ingest.Result, error) {
	t.Helper()
	return func(res ingest.Result, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("load %s: %v", index, err)
		}
		if res.Failed > 0 || res.Count != want {
			t.Fatalf("load %s: sent %d, failed %d, count %d, want %d", index, res.Sent, res.Failed, res.Count, want)
		}
	}
}
