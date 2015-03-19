package main

import (
	"database/sql"
	"flag"
	"fmt"
	ical "github.com/fredcy/icalendar"
	"github.com/fredcy/psfacade"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const userprefix = "/pscal/u/"
const roomprefix = "/pscal/r/"
const calprefix = "/pscal/cal"

var address = flag.String("address", ":8080", "Listen and serve at this address")
var logflags = flag.Int("logflags", 3, "Flags to standard logger")
var maxage = flag.Int("maxage", 8*3600, "Cache-Control max-age value")

var dsn string
var dsnre = regexp.MustCompile(`^(.*?)/(.*?)@(.*?):(.*)`)

// set_dsn sets the DSN for accessing the PowerSchool database.
func set_dsn() {
	dsn = os.Getenv("PS_DSN")
	match := dsnre.FindStringSubmatch(dsn)
	if match == nil {
		log.Panic("DSN value is not well formed:", dsn)
	}
	pshost := match[3]
	log.Printf("PowerSchool host is %s", pshost)
}

func calhandler(generator func(*http.Request) *ical.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		starttime := time.Now()
		w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", *maxage))
		w.Header().Set("Content-Type", "text/calendar")
		w.Header().Set("Last-Modified", starttime.Format("Mon, 02 Jan 2006 15:04:05 MST"))

		cal := generator(r)
		w.Write([]byte(cal.String()))

		client := r.RemoteAddr
		forwarded_for := strings.Join(r.Header["X-Forwarded-For"], "")
		if forwarded_for != "" {
			client += " (" + forwarded_for + ")"
		}
		endtime := time.Now()
		log.Printf("served %v (%d components) to %v in %v",
			r.URL, cal.ComponentCount(), client, endtime.Sub(starttime))
	}
}

func usergenerator(r *http.Request) *ical.Component {
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		log.Panicf("Cannot open database: %s", err)
	}
	defer db.Close()
	loginid := r.URL.Path[len(userprefix):]
	return psfacade.TeacherCalendar(db, loginid)
}

func roomgenerator(r *http.Request) *ical.Component {
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		log.Panicf("Cannot open database: %s", err)
	}
	defer db.Close()
	roomname := r.URL.Path[len(roomprefix):]
	return psfacade.RoomCalendar(db, roomname)
}

func maingenerator(r *http.Request) *ical.Component {
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		log.Panicf("Cannot open database: %s", err)
	}
	defer db.Close()
	return psfacade.GetCalendar(db)
}

func main() {
	flag.Parse()
	log.SetFlags(*logflags)
	set_dsn()

	http.HandleFunc(userprefix, calhandler(usergenerator))
	http.HandleFunc(roomprefix, calhandler(roomgenerator))
	http.HandleFunc(calprefix, calhandler(maingenerator))

	log.Printf("Listening at %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
