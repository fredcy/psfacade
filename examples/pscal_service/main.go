package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"github.com/fredcy/psfacade"
	"os"
	"time"
)

const prefix = "/pscal/"

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
	w.Header().Set("Cache-Control", "public,max-age=3600")
	w.Write([]byte(cal))
	endtime := time.Now()
	log.Printf("served %v to %v in %v", r.URL, r.RemoteAddr, endtime.Sub(starttime))
}

func main() {
	var err error
	flag.Parse()
	log.SetFlags(*logflags)
	user := os.Getenv("PS_USER")
	password := os.Getenv("PS_PASSWORD")
	host := os.Getenv("PS_HOST")
	dsn = fmt.Sprintf("%s/%s@//%s:1521/PSProdDB", user, password, host)

	http.HandleFunc(prefix, handler)
	log.Printf("PowerSchool host is %s", host)
	log.Printf("Listening on port %s", *port)
	err = http.ListenAndServe(":" + *port, nil)
	if err != nil {
		log.Panic(err)
	}
}
