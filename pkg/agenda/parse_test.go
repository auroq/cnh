package agenda

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"nothing", "", nil},
		{"only blank lines", "\n  \n\t\n", nil},
		{"three titles one per line", "Sign Language Crash Course\nKeynote: Welcome\nSponsored Keynote", []string{"Sign Language Crash Course", "Keynote: Welcome", "Sponsored Keynote"}},
		{"titles with surrounding spaces and Windows line ends", "  Sign Language Crash Course \r\nKeynote: Welcome\r\n", []string{"Sign Language Crash Course", "Keynote: Welcome"}},
		{"the same title twice", "Keynote: Welcome\nKeynote: Welcome", []string{"Keynote: Welcome"}},
		{"an iCal export", strings.Join([]string{
			"BEGIN:VCALENDAR",
			"VERSION:2.0",
			"BEGIN:VEVENT",
			"DTSTART:20261111T153000Z",
			"SUMMARY:Sign Language Crash Course",
			"LOCATION:Grand Ballroom J",
			"END:VEVENT",
			"BEGIN:VEVENT",
			"SUMMARY:Keynote: Welcome\\, Everyone",
			"END:VEVENT",
			"END:VCALENDAR",
		}, "\r\n"), []string{"Sign Language Crash Course", "Keynote: Welcome, Everyone"}},
		{"an iCal export with a folded summary", strings.Join([]string{
			"BEGIN:VCALENDAR",
			"BEGIN:VEVENT",
			"SUMMARY:Sign Language",
			"  Crash Course",
			"END:VEVENT",
			"END:VCALENDAR",
		}, "\r\n"), []string{"Sign Language Crash Course"}},
		{"an iCal export with a language parameter on the summary", "BEGIN:VCALENDAR\nSUMMARY;LANGUAGE=en-us:Keynote\nEND:VCALENDAR", []string{"Keynote"}},
	}
	for _, c := range cases {
		t.Run("when the paste is "+c.name, func(t *testing.T) {
			got := Parse(c.in)
			t.Run("it should read "+describe(c.want), func(t *testing.T) {
				if strings.Join(got, "|") != strings.Join(c.want, "|") {
					t.Fatalf("got %q, want %q", got, c.want)
				}
			})
		})
	}
}

func describe(titles []string) string {
	if len(titles) == 0 {
		return "no titles"
	}
	return strings.Join(titles, " and ")
}
