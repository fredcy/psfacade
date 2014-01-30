package psfacade

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
	ical "github.com/fredcy/icalendar"
)

// Meeting holds all PowerSchool data for a single teacher schedule
// event, a course meeting.
type Meeting struct {
	loginid string
	start time.Time
	duration int
	course_name string
	course_number string
	section_number int
	room string
}

// GetTeacherSched returns a channel of Meeting items for the given teacher username.
func GetTeacherSched(db *sql.DB, name string) <-chan Meeting {
	query := `
    with
    sm1 as (select sm.sectionid, sm.cycle_day_letter, min(sm.period_number) period_min from section_meeting sm group by sectionid, cycle_day_letter),
    sm2 as (select sm.sectionid, sm.cycle_day_letter, max(sm.period_number) period_max from section_meeting sm group by sectionid, cycle_day_letter)
    select
    teachers.loginid,
    to_char(cd.date_value, 'YYYYMMDD') "date",
    to_char(floor(bsi1.start_time/3600), 'FM09')
    || to_char(floor(mod(bsi1.start_time, 3600) / 60), 'FM09') "start", -- HHMM
    floor((bsi2.end_time - bsi1.start_time) / 60) duration, -- minutes
    courses.course_name,
    s.course_number,
    s.section_number,
    s.room
    from sections s
    join teachers on s.teacher = teachers.id
    join courses on s.course_number = courses.course_number
    join sm1 on s.id = sm1.sectionid
    join sm2 on s.id = sm2.sectionid and sm1.cycle_day_letter = sm2.cycle_day_letter
    join terms on s.termid = terms.id and s.schoolid = terms.schoolid
    join period period1 on sm1.period_min = period1.period_number and s.schoolid = period1.schoolid and terms.yearid = period1.year_id
    join period period2 on sm2.period_max = period2.period_number and s.schoolid = period2.schoolid and terms.yearid = period2.year_id
    join cycle_day on sm1.cycle_day_letter = cycle_day.letter and terms.yearid = cycle_day.year_id and cycle_day.schoolid = terms.schoolid
    -- up to here we've got one row per section meeting:  e.g. MAT321-1 A(13-15)
    join calendar_day cd on cd.schoolid = s.schoolid and cd.date_value between terms.firstday and terms.lastday and cd.cycle_day_id = cycle_day.id
    -- now we've matched the section meetings against each calendar day they could meet (if bell sched allows)
    join bell_schedule_items bsi1 on period1.id = bsi1.period_id and cd.bell_schedule_id = bsi1.bell_schedule_id
    join bell_schedule_items bsi2 on period2.id = bsi2.period_id and cd.bell_schedule_id = bsi2.bell_schedule_id
    -- matched against bell schedule to determine if that day has the periods, and get the actual period times
    where
    s.schoolid = 140177
    and terms.yearid = :yearid
    and teachers.loginid = :loginid
    and period1.period_number < 21
    and s.course_number not in ('SLD100', 'SLD102', 'SLD200', 'SLD210', 'SLD600')  -- Res Life, LASSI, Nav, LEAD, I-Day Attendance
    and teachers.loginid is not null  -- ignore placeholders like "Staff, New"
    order by teachers.loginid, cd.date_value, sm1.period_min
`
	yearid := get_yearid()
	if os.Getenv("TEACHER_SCHED_DEBUG") != "" {
		log.Printf("yearid=%v, name=%v, query=%v", yearid, name, query)
	}

	rows, err := db.Query(query, yearid, name)
	if err != nil {
		log.Panicf("query failed: %v", err)
	}
	ch := make(chan Meeting)
	go func() {
		defer rows.Close()		// must be inside goroutine so we don't close until done
		var (
			date string
			start string
		)
		loc, err:= time.LoadLocation("America/Chicago")
		if err != nil {
			log.Panicf("LoadLocation failed: %v", err)
		}
		for rows.Next() {
			m := Meeting{}
			err = rows.Scan(&m.loginid, &date, &start, &m.duration, &m.course_name, &m.course_number, &m.section_number, &m.room)
			if err != nil {
				log.Panicf("rows.Scan(): %v", err)
			}
			datetimestr := date + start
			m.start, err = time.ParseInLocation("200601021504", datetimestr, loc)
			if err != nil {
				log.Panic("time.Parse(): %v", err)
			}
			//log.Printf("m = %v, m.start = %v, datetimestr = %v", m, m.start, datetimestr)
			ch <- m
		}
		close(ch)
	}()
	return ch
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

// timezone is the constant VTIMEZONE element for our calendars
var Timezone = cal_timezone()

func TeacherCalendar(db *sql.DB, loginid string) *ical.Component {
	ch := GetTeacherSched(db, loginid)
	cal := ical.Component{}
	cal.SetName("VCALENDAR")
	cal.Set("VERSION", ical.VString("2.0"))
	cal.Set("PRODID", ical.VStringf("-//imsa.edu//powerschool calendar for %s//EN", loginid))
	cal.Set("METHOD", ical.VString("PUBLISH"))
	cal.Set("CALSCALE", ical.VString("GREGORIAN"))
	cal.Set("x-wr-calname", ical.VStringf("%s@imsa.edu PowerSchool", loginid))
	cal.Set("x-wr-caldesc", ical.VStringf("IMSA PowerSchool teacher calendar for %s", loginid))
	cal.Set("x-wrt-timezone", ical.VString("America/Chicago"))
	cal.AddComponent(&Timezone)
		
	for mtg := range ch {
		e := ical.Component{}
		e.SetName("VEVENT")
		dtstart := ical.VDateTime(mtg.start)
		e.Set("DTSTART", dtstart)
		e.Set("DURATION", ical.VDuration(time.Duration(mtg.duration)*time.Minute))
		e.Set("SUMMARY", ical.VString(mtg.course_name))
		e.Set("DESCRIPTION", ical.VString(fmt.Sprintf("%s (%s-%d) -- %s",
			mtg.course_name, mtg.course_number, mtg.section_number, mtg.room)))
		organizer := ical.NewProperty("ORGANIZER", ical.VString(fmt.Sprintf("mailto:%s@imsa.edu", mtg.loginid)))
		//organizer.AddParameter("CN", ical.VString("TODO-CN"))
		e.AddProperty(organizer)
		e.Set("DTSTAMP", ical.VDateTime(time.Now()))
		e.Set("UID", ical.VString(fmt.Sprintf("%s-TODO@imsa.edu", dtstart.String())))
		attendee := ical.NewProperty("ATTENDEE", ical.VString(fmt.Sprintf("mailto:%s@imsa.edu", mtg.loginid)))
		attendee.AddParameter("PARTSTAT", ical.VString("ACCEPTED"))
		attendee.AddParameter("ROLE", ical.VString("REQ-PARTICIPANT"))
		//attendee.AddParameter("CN", ical.VString("TODO-CN"))
		// TODO: add attendee for room
		e.AddProperty(attendee)
		cal.AddComponent(&e)
	}
	return &cal
}
