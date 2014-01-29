package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"github.com/fredcy/psfacade"
	"os"
	"regexp"
	"strings"
	"time"
)

const prefix = "/pscal/u/"

var address = flag.String("address", ":8080", "Listen and serve at this address")
var logflags = flag.Int("logflags", 3, "Flags to standard logger")
var maxage = flag.Int("maxage", 8*3600, "Cache-Control max-age value")

var dsn string
var dsnre = regexp.MustCompile(`^(.*?)/(.*?)@//(.*?):(.*)`)

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

// calhander responds with the icalendar data for the teacher whose
// username is given in the request URL path.
func calhandler(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	loginid := r.URL.Path[len(prefix):]
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		log.Panicf("Cannot open database: %s", err)
	}
	defer db.Close()
	cal := psfacade.TeacherCalendar(db, loginid)
	w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d", *maxage))
	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Last-Modified", starttime.Format("Mon, 02 Jan 2006 15:04:05 MST"))
	w.Write([]byte(cal.String()))
	endtime := time.Now()

	client := r.RemoteAddr
	forwarded_for := strings.Join(r.Header["X-Forwarded-For"], "")
	if forwarded_for != "" {
		client += " (" + forwarded_for + ")"
	}
	log.Printf("served %v (%d components) to %v in %v",
		r.URL, cal.ComponentCount(), client, endtime.Sub(starttime))
}

func main() {
	flag.Parse()
	log.SetFlags(*logflags)
	set_dsn()

	http.HandleFunc(prefix, calhandler)

	log.Printf("Listening at %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
