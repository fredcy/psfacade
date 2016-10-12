package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fredcy/psfacade"
	"github.com/gorilla/mux"
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintln(w, "[")
	enc := json.NewEncoder(w)

	first := true
	for s := range students {
		if !first {
			fmt.Fprintf(w, ",")
		}

		if err := enc.Encode(&s); err != nil {
			log.Println(err)
			return
		}
		first = false
	}

	fmt.Fprintln(w, "]")
}

var address = flag.String("address", ":8080", "Listen and serve at this address")

func main() {
	flag.Parse()

	dsn := os.Getenv("PS_DSN")
	db, err := sql.Open("oci8", dsn)
	if err != nil {
		log.Panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/students", wraptimer(wrapdb(studentshandler, db)))
	http.Handle("/", &MyServer{r})

	log.Printf("Listening at %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}

/* See http://stackoverflow.com/questions/12830095/setting-http-headers-in-golang about CORS */

type MyServer struct {
	r *mux.Router
}

func (s *MyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	// Lets Gorilla work
	s.r.ServeHTTP(rw, req)
}
