package psfacade

import (
	"database/sql"
	//"fmt"
	"log"
	"os"
	//ical "github.com/fredcy/icalendar"
	"time"
)

type CalDay struct {
	date time.Time
	insession int
	note string
	bell_sched string
	cycle_day string
}

func GetCalendar(db *sql.DB) {
	query := `
--SELECT cast(cd.date_value as timestamp) date_value, cd.insession, cd.note, bs.name, cyd.abbreviation
SELECT to_char(cd.date_value, 'IYYY-MM-DD') date_str, cd.insession, cd.note, bs.name, cyd.abbreviation
--SELECT cd.insession, cd.note, bs.name, cyd.abbreviation
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
	defer rows.Close()

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
	}
	err = rows.Err()
	if err != nil {
		log.Panic("rows.Err ", err)
	}
}
