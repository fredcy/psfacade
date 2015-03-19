package psfacade

import (
	"database/sql"
	"fmt"
	ical "github.com/fredcy/icalendar"
	"log"
	"os"
	"strings"
	"time"
)

// CalDay is a single PowerSchool calendar day
type CalDay struct {
	date       time.Time
	insession  int
	note       string
	bell_sched string
	cycle_day  string
}

func emptyifnull(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

// GetCalendarDays returns a channel with all of the calendar items
func GetCalendarDays(db *sql.DB) <-chan CalDay {
	query := `
SELECT to_char(cd.date_value, 'IYYY-MM-DD') date_str, cd.insession, cd.note, bs.name, cyd.abbreviation
FROM terms terms1
join terms terms2 on terms1.schoolid = terms2.schoolid
join calendar_day cd on cd.date_value between terms1.firstday and terms2.lastday and cd.schoolid = terms1.schoolid
left outer join bell_schedule bs on cd.bell_schedule_id = bs.id
left outer join cycle_day cyd on cd.cycle_day_id = cyd.id
where terms1.id = :termid1 and terms2.id = :termid2 and terms1.schoolid = 140177
`
	yearid := get_yearid()
	termid2 := yearid * 100
	termid1 := (yearid - 1) * 100
	debug := os.Getenv("CALENDAR_DEBUG") != ""
	if debug {
		log.Println("termid", termid1, "query", query)
	}

	rows, err := db.Query(query, termid1, termid2)
	if err != nil {
		log.Panicf("query failed: %v", err)
	}
	days := make(chan CalDay)
	go func() {
		defer rows.Close()
		defer close(days)
		for rows.Next() {
			cd := CalDay{}
			var date string
			var note, bell_sched, cycle_day sql.NullString
			err = rows.Scan(&date, &cd.insession, &note, &bell_sched, &cycle_day)
			if err != nil {
				log.Panic("rows.Scan: ", err)
			}
			cd.date, err = time.Parse("2006-01-02", date)
			if err != nil {
				log.Panic("time.Parse ", err)
			}
			cd.note = emptyifnull(note)
			cd.bell_sched = emptyifnull(bell_sched)
			cd.cycle_day = emptyifnull(cycle_day)
			if debug {
				log.Printf("date=%v insession=%v note='%v' bell_sched='%v' cycle_day='%v'",
					cd.date, cd.insession, cd.note, cd.bell_sched, cd.cycle_day)
			}
			days <- cd
		}
		err = rows.Err()
		if err != nil {
			log.Panic("rows.Err ", err)
		}
	}()
	return days
}

// GetCalendar returns iCalendar data for PowerSchool common calendar (ABCDI days, bell schedules, notes)
func GetCalendar(db *sql.DB) *ical.Component {
	days := GetCalendarDays(db)
	cal := ical.Component{}
	cal.SetName("VCALENDAR")
	cal.Set("VERSION", ical.VString("2.0"))
	cal.Set("PRODID", ical.VStringf("-//imsa.edu//powerschool calendar//EN"))
	cal.Set("METHOD", ical.VString("PUBLISH"))
	cal.Set("CALSCALE", ical.VString("GREGORIAN"))
	cal.Set("x-wr-calname", ical.VStringf("IMSA PowerSchool"))
	cal.Set("x-wr-caldesc", ical.VStringf("IMSA PowerSchool common calendar"))
	cal.Set("x-wrt-timezone", ical.VString("America/Chicago"))
	vtimezone := cal_timezone()
	cal.AddComponent(&vtimezone)

	dtstamp := ical.VDateTime(time.Now())
	for day := range days {
		summary := format_summary(&day)
		if summary == "" {
			continue
		}
		e := ical.Component{}
		e.SetName("VEVENT")
		e.Set("DTSTART", ical.VDate(day.date)).Add("VALUE", ical.VString("DATE"))
		e.Set("DTEND", ical.VDate(day.date.AddDate(0, 0, 1))).Add("VALUE", ical.VString("DATE"))
		// this pattern of start and end makes the event an all-day event that displays at top
		e.Set("SUMMARY", ical.VString(summary))
		e.Set("DESCRIPTION", ical.VString(format_description(&day)))
		e.Set("DTSTAMP", dtstamp)
		e.Set("UID", ical.VString(fmt.Sprintf("PS-Calendar-%s@imsa.edu", day.date.Format("20060102"))))
		cal.AddComponent(&e)
	}
	return &cal
}

var cycle_day_display = map[string]bool{
	"A": true,
	"B": true,
	"C": true,
	"D": true,
	"I": true,
}

// Generate SUMMARY string for given calendar item
func format_summary(day *CalDay) string {
	var summary string
	if cycle_day_display[day.cycle_day] {
		summary += day.cycle_day
		if day.bell_sched != "" && !strings.HasPrefix(day.bell_sched, "Full Day") {
			summary += fmt.Sprintf(" (%s)", day.bell_sched)
		}
	} else {
		//log.Printf("ignoring cycle day %v", day.cycle_day)
	}
	if day.note != "" {
		if summary != "" {
			summary += ": "
		}
		summary += day.note
	}
	return summary
}

// Generate DESCRIPTION string for given calendar item
func format_description(day *CalDay) string {
	var description string
	if day.cycle_day != "" {
		description += ("Cycle Day: " + day.cycle_day + "\n")
	}
	if day.bell_sched != "" {
		description += ("Bell Schedule: " + day.bell_sched + "\n")
	}
	if day.note != "" {
		description += ("Note: " + day.note + "\n")
	}
	return description
}
