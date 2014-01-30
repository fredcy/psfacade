package psfacade

import (
	"time"
	ical "github.com/fredcy/icalendar"
)

func get_yearid() int {
	now := time.Now()
	var academicyear int
	if now.Month() < 7 {
		academicyear = now.Year()
	} else {
		academicyear = now.Year() + 1
	}
	yearid := academicyear - 1991 // the usual PowerSchool conversion
	return yearid
}

// cal_timezone returns a standard VTIMEZONE element for the America/Chicago zone
func cal_timezone() ical.Component {
	timezone := ical.Component{}
	timezone.SetName("VTIMEZONE")
	timezone.Set("TZID", ical.VString("America/Chicago"))
	daylight := ical.Component{}
	daylight.SetName("DAYLIGHT")
	daylight.Add("tzname", ical.VString("CDT"))
	daylight.Add("tzoffsetfrom", ical.VUtcOffset(-6*3600))
	daylight.Add("tzoffsetto", ical.VUtcOffset(-5*3600))
	daylight.Add("dtstart", ical.VDateTime(time.Date(1970, 3, 8, 2, 0, 0, 0, time.UTC)))
	rrv := ical.VEnumList{}
	rrv.AddValue("FREQ", ical.VString("YEARLY"))
	rrv.AddValue("BYMONTH", ical.VInt(3))
	rrv.AddValue("BYDAY", ical.VString("2SU"))
	daylight.Add("RRULE", rrv)
	timezone.AddComponent(&daylight)
	standard := ical.Component{}
	standard.SetName("STANDARD")
	standard.Add("tzname", ical.VString("CST"))
	standard.Add("tzoffsetfrom", ical.VUtcOffset(-5*3600))
	standard.Add("tzoffsetto", ical.VUtcOffset(-6*3600))
	standard.Add("dtstart", ical.VDateTime(time.Date(1970, 11, 1, 2, 0, 0, 0, time.UTC)))
	rrv = ical.VEnumList{}
	rrv.AddValue("FREQ", ical.VString("YEARLY"))
	rrv.AddValue("BYMONTH", ical.VInt(11))
	rrv.AddValue("BYDAY", ical.VString("1SU"))
	standard.Add("RRULE", rrv)
	timezone.AddComponent(&standard)
	return timezone
}
