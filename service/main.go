package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/fredcy/psfacade"
	"log"
	"net/http"
	"os"
	"time"
)

type dbfunc func(http.ResponseWriter, *http.Request, *sql.DB)

func wrapdb(fn dbfunc, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db)
	}
}

func wraptimer(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		starttime := time.Now()
		fn(w, r)
		endtime := time.Now()
		log.Printf("Served %v in %v", r.URL, endtime.Sub(starttime))
	}
}

func studentshandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	students := psfacade.GetStudents(db)
	for s := range students {
		_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.Number, s.FirstName, s.LastName, s.Room, s.Birthdate)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

var address = flag.String("address", ":8080", "Listen and serve at this address")

func main() {
	flag.Parse()
	dsn := os.Getenv("PS_DSN")
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		log.Panic(err)
	}
	http.HandleFunc("/students", wraptimer(wrapdb(studentshandler, db)))
	log.Printf("Listening at %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}