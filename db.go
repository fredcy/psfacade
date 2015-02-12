package psfacade

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	_ "github.com/mattn/go-oci8"
)

type Configuration struct {
	Server	string
	Port	string
	Database string
	Uid string
	Pwd string
}

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

func MakeDSN(config Configuration) string {
	return fmt.Sprintf("%s/%s@%s:%s/%s", config.Uid, config.Pwd, config.Server, config.Port, config.Database)
}

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
	//defer rows.Close()

	return rows, nil
}
