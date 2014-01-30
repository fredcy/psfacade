package psfacade

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	ical "github.com/fredcy/icalendar"
	"strings"
	"time"
)

type CalDay struct {
	date time.Time
	insession int
	note string
	bell_sched string
	cycle_day string
}

func GetCalendarDays(db *sql.DB) <-chan CalDay {
	query := `
SELECT to_char(cd.date_value, 'IYYY-MM-DD') date_str, cd.insession, cd.note, bs.name, cyd.abbreviation
FROM   terms
join calendar_day cd on cd.date_value between terms.firstday and terms.lastday and cd.schoolid = terms.schoolid
left outer join bell_schedule bs on cd.bell_schedule_id = bs.id
left outer join cycle_day cyd on cd.cycle_day_id = cyd.id
where terms.id = :termid and terms.schoolid = 140177
`
	yearid := get_yearid()
	termid := yearid * 100
	debug := os.Getenv("CALENDAR_DEBUG") != ""
	if debug { log.Println("termid", termid, "query", query) }

	rows, err := db.Query(query, termid)
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
			err = rows.Scan(&date, &cd.insession, &cd.note, &cd.bell_sched, &cd.cycle_day)
			if err != nil {
				log.Panic("rows.Scan", err)
			}
			cd.date, err = time.Parse("2006-01-02", date)
			if err != nil {
				log.Panic("time.Parse ", err)
			}
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
	duration := ical.VDuration(time.Duration(time.Hour * 24)) // all events are full day
	for day := range days {
		summary := format_summary(&day)
		if summary == "" {
			continue
		}
		e := ical.Component{}
		e.SetName("VEVENT")
		dtstart := ical.VDateTime(day.date)
		e.Set("DTSTART", dtstart)
		e.Set("DURATION", duration)
		e.Set("SUMMARY", ical.VString(summary))
		e.Set("DESCRIPTION", ical.VString(format_description(&day)))
		e.Set("DTSTAMP", dtstamp)
		e.Set("UID", ical.VString(fmt.Sprintf("PS-Calendar-%v@imsa.edu", dtstart)))
		cal.AddComponent(&e)
	}
	return &cal
}

var cycle_day_display = map[string]bool {
	"A": true,
	"B": true,
	"C": true,
	"D": true,
	"I": true,
}

func format_summary(day *CalDay) string {
	var summary string
	if cycle_day_display[day.cycle_day] {
		summary += day.cycle_day
		if day.bell_sched != "" && ! strings.HasPrefix(day.bell_sched, "Full Day")  {
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

