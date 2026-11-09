package acceptance

type distance int

const (
	noDistance distance = iota
	sameLevel
	sameBuilding
)

func (d distance) String() string {
	switch d {
	case sameLevel:
		return "it is on the same level"
	case sameBuilding:
		return "it is in the same building on another level"
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
		{From: "", Place: "quiet-room-260a", Distance: noDistance, Verified: true},
	}},
	{"quiet room", []expectation{
		{From: "Grand Ballroom J", Place: "quiet-room-260a", Distance: sameBuilding, Verified: true},
	}},
	{"nearest elevator", []expectation{
		{From: "Grand Ballroom J", Place: "elevators-salt-palace", Distance: sameBuilding, Verified: false},
		{From: "Salt Lake Ballroom C", Place: "elevators-hyatt", Distance: sameBuilding, Verified: false},
		{From: "Regency Ballroom B-D", Place: "elevators-hyatt", Distance: sameBuilding, Verified: false},
		{From: "Capitol Reef", Place: "elevators-hyatt", Distance: sameBuilding, Verified: false},
		{From: "", Place: "elevators-salt-palace", Distance: noDistance, Verified: false},
	}},
	{"step-free entrance", []expectation{
		{From: "", Place: "step-free-200-south", Distance: noDistance, Verified: false},
	}},
	{"prayer room", []expectation{
		{From: "Regency Ballroom B-D", Place: "prayer-room-zion", Distance: sameBuilding, Verified: true},
		{From: "Capitol Reef", Place: "prayer-room-zion", Distance: sameLevel, Verified: true},
		{From: "", Place: "prayer-room-zion", Distance: noDistance, Verified: true},
	}},
	{"nursing room", []expectation{
		{From: "Grand Ballroom J", Place: "nursing-room", Distance: sameBuilding, Verified: true},
		{From: "", Place: "nursing-room", Distance: noDistance, Verified: true},
	}},
	{"coat and bag check", []expectation{
		{From: "255 BC", Place: "coat-check-east", Distance: sameLevel, Verified: true},
		{From: "", Place: "coat-check-east", Distance: noDistance, Verified: true},
	}},
	{"registration and help desk", []expectation{
		{From: "", Place: "registration-200-south", Distance: noDistance, Verified: true},
	}},
}
