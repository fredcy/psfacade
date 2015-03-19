package psfacade

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-oci8" // needed to define "oci8" driver
	"log"
	"os"
)

// Configuration defines the PowerSchool database connection
type Configuration struct {
	Server   string
	Port     string
	Database string
	Uid      string
	Pwd      string
}

// GetConfig reads the config file and returns the PowerSchool connection data
func GetConfig(filename string) Configuration {
	conffile, err := os.Open(filename)
	if err != nil {
		log.Panicf("cannot open config file (%v)", filename)
	}
	log.Printf("Reading %s for Oracle config", filename)
	decoder := json.NewDecoder(conffile)
	configuration := Configuration{}
	jerr := decoder.Decode(&configuration)
	if jerr != nil {
		log.Panicf("Cannot decode json file %v: %v", filename, jerr)
	}
	return configuration
}

// MakeDSN generates an oci8 DSN value from the given Configuration
func MakeDSN(config Configuration) string {
	return fmt.Sprintf("%s/%s@%s:%s/%s", config.Uid, config.Pwd, config.Server, config.Port, config.Database)
}

// RunQuery opens the connection defined by the Configuration, runs the query
// (passing any args), and returns the rows.
func RunQuery(config Configuration, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := sql.Open("oci8", MakeDSN(config))
	if err != nil {
		log.Printf("sql.Open error: %v", err)
		return &sql.Rows{}, err
	}
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("db.Query error (query='%v'): %v", query, err)
		return &sql.Rows{}, err
	}

	return rows, nil
}
