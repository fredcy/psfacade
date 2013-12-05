package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"github.com/fredcy/psfacade"
	"time"
)

const prefix = "/pscal/"

var conffilename = flag.String("conf", "ps.conf", "PowerSchool database connection config file")
var port = flag.String("port", "8080", "Listen and serve on this port")
var logflags = flag.Int("logflags", 3, "Flags to standard logger")

var dsn string

var slog *log.Logger

func handler(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	loginid := r.URL.Path[len(prefix):]
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		log.Panicf("Cannot open database: %s", err)
	}
	defer db.Close()
	cal := psfacade.TeacherCalendar(db, loginid)
	w.Write([]byte(cal))
	endtime := time.Now()
	log.Printf("served %v to %v in %v", r.URL, r.RemoteAddr, endtime.Sub(starttime))
}

func main() {
	var err error
	flag.Parse()
	log.SetFlags(*logflags)
	config := psfacade.GetConfig(*conffilename)
	dsn = psfacade.MakeDSN(config)

	http.HandleFunc(prefix, handler)
	log.Printf("Listening on port %s", *port)
	err = http.ListenAndServe(":" + *port, nil)
	if err != nil {
		log.Panic(err)
	}
}
