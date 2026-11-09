package web

import (
	"fmt"

	"github.com/auroq/cnh/pkg/ask"
)

const soon = 120

var distanceWords = map[string]string{
	ask.DistanceSameLevel:     "Same level as you",
	ask.DistanceSameBuilding:  "Same building, another level",
	ask.DistanceOtherBuilding: "Other building",
}

func when(s ask.Session) string {
	switch {
	case s.State == ask.SessionRunning:
		return "Happening now, ends " + hour(s.End)
	case s.State == ask.SessionEnded:
		return "Ended at " + hour(s.End)
	case s.StartsInMinutes == 1:
		return "Starts in 1 minute"
	case s.StartsInMinutes <= soon:
		return fmt.Sprintf("Starts in %d minutes", s.StartsInMinutes)
	default:
		return "Starts " + s.Start.In(mountain).Format("Monday 3:04 pm")
	}
}

func distance(tier string) string {
	return distanceWords[tier]
}
