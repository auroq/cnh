package acceptance

type distance int

const (
	noDistance distance = iota
	sameLevel
	sameBuilding
	otherBuilding
)

func (d distance) String() string {
	switch d {
	case sameLevel:
		return "it is on the same level"
	case sameBuilding:
		return "it is in the same building on another level"
	case otherBuilding:
		return "it is in the other building"
	default:
		return "nothing about the distance"
	}
}

type expectation struct {
	From     string
	Place    string
	Distance distance
	Verified bool
}

type question struct {
	Ask  string
	Rows []expectation
}

type needCase struct {
	Say  string
	Need string
	Kind string
	Note string
}

var needCases = []needCase{
	{"I can't do stairs", "a step-free way around", "step-free-entrance", ""},
	{"I need to fix my makeup", "somewhere private", "all-gender-restroom", "Single-stall where known."},
	{"where can I inject insulin", "a private place for medication", "nursing-room", "first aid staff can help"},
	{"I need a tampon", "period products", "all-gender-restroom", "Ask at the information desk if the restroom has none."},
	{"I'm hurt", "first aid", "info-desk", "Staff at registration know where the first-aid points are."},
	{"emergency exit", "the way out", "step-free-entrance", "follow the lit exit signs"},
	{"I'm overwhelmed", "somewhere to rest", "quiet-room", ""},
	{"I'm lost", "help from staff", "info-desk", ""},
	{"I can't walk up the stairs", "a step-free way around", "step-free-entrance", ""},
}

var cases = []question{
	{"nearest all-gender restroom", []expectation{
		{From: "255 BC", Place: "agr-sp-l2-255", Distance: sameLevel, Verified: true},
		{From: "Grand Ballroom J", Place: "agr-sp-l1-showcase", Distance: sameLevel, Verified: true},
		{From: "Salt Lake Ballroom C", Place: "agr-hy-l2-juniper", Distance: sameLevel, Verified: true},
		{From: "Park City", Place: "agr-hy-l3-powder", Distance: sameLevel, Verified: true},
		{From: "Regency Ballroom B-D", Place: "agr-hy-l4-broadcast", Distance: sameLevel, Verified: true},
		{From: "Narnia", Place: "agr-sp-l2-255", Distance: noDistance, Verified: true},
		{From: "", Place: "agr-sp-l2-255", Distance: noDistance, Verified: true},
	}},
	{"nearest quiet room", []expectation{
		{From: "255 BC", Place: "quiet-room-260a", Distance: sameLevel, Verified: true},
		{From: "Grand Ballroom J", Place: "quiet-room-260a", Distance: sameBuilding, Verified: true},
		{From: "Park City", Place: "quiet-room-260a", Distance: otherBuilding, Verified: true},
		{From: "", Place: "quiet-room-260a", Distance: noDistance, Verified: true},
	}},
	{"quiet room", []expectation{
		{From: "Grand Ballroom J", Place: "quiet-room-260a", Distance: sameBuilding, Verified: true},
	}},
	{"nearest elevator", []expectation{
		{From: "255 BC", Place: "elevators-salt-palace", Distance: sameBuilding, Verified: false},
		{From: "Grand Ballroom J", Place: "elevators-salt-palace", Distance: sameBuilding, Verified: false},
		{From: "Park City", Place: "elevators-hyatt", Distance: sameBuilding, Verified: false},
		{From: "Salt Lake Ballroom C", Place: "elevators-hyatt", Distance: sameBuilding, Verified: false},
		{From: "Regency Ballroom B-D", Place: "elevators-hyatt", Distance: sameBuilding, Verified: false},
		{From: "Capitol Reef", Place: "elevators-hyatt", Distance: sameBuilding, Verified: false},
		{From: "", Place: "elevators-salt-palace", Distance: noDistance, Verified: false},
	}},
	{"step-free entrance", []expectation{
		{From: "255 BC", Place: "step-free-200-south", Distance: sameBuilding, Verified: false},
		{From: "Grand Ballroom J", Place: "step-free-200-south", Distance: sameBuilding, Verified: false},
		{From: "Salt Lake Ballroom C", Place: "step-free-200-south", Distance: otherBuilding, Verified: false},
		{From: "", Place: "step-free-200-south", Distance: noDistance, Verified: false},
	}},
	{"prayer room", []expectation{
		{From: "255 BC", Place: "prayer-room-zion", Distance: otherBuilding, Verified: true},
		{From: "Grand Ballroom J", Place: "prayer-room-zion", Distance: otherBuilding, Verified: true},
		{From: "Salt Lake Ballroom C", Place: "prayer-room-zion", Distance: sameBuilding, Verified: true},
		{From: "Park City", Place: "prayer-room-zion", Distance: sameBuilding, Verified: true},
		{From: "Regency Ballroom B-D", Place: "prayer-room-zion", Distance: sameBuilding, Verified: true},
		{From: "Capitol Reef", Place: "prayer-room-zion", Distance: sameLevel, Verified: true},
		{From: "", Place: "prayer-room-zion", Distance: noDistance, Verified: true},
	}},
	{"nursing room", []expectation{
		{From: "255 BC", Place: "nursing-room", Distance: sameBuilding, Verified: true},
		{From: "Grand Ballroom J", Place: "nursing-room", Distance: sameBuilding, Verified: true},
		{From: "Park City", Place: "nursing-room", Distance: otherBuilding, Verified: true},
		{From: "", Place: "nursing-room", Distance: noDistance, Verified: true},
	}},
	{"coat and bag check", []expectation{
		{From: "255 BC", Place: "coat-check-east", Distance: sameLevel, Verified: true},
		{From: "Grand Ballroom J", Place: "coat-check-east", Distance: sameBuilding, Verified: true},
		{From: "Park City", Place: "coat-check-east", Distance: otherBuilding, Verified: true},
		{From: "", Place: "coat-check-east", Distance: noDistance, Verified: true},
	}},
	{"registration and help desk", []expectation{
		{From: "255 BC", Place: "registration-200-south", Distance: sameBuilding, Verified: true},
		{From: "Park City", Place: "registration-200-south", Distance: otherBuilding, Verified: true},
		{From: "", Place: "registration-200-south", Distance: noDistance, Verified: true},
	}},
}
