package psfacade

import (
	"database/sql"
	"fmt"
	_ "log"
	"testing"
)

func TestSimpleQuery(t *testing.T) {
	config := GetConfig("ps.conf")
	query := "select 3.14 a, 'foo' b from dual"
	rows, err := RunQuery(config, query)
	if err != nil {
		t.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	query2 := "SELECT student_number from Students where last_name = :1"
	rows2, err2 := RunQuery(config, query2, "Yan")
	if err2 != nil {
		t.Errorf("query2 failed: %v", err2)
	}
	defer rows2.Close()

	var sns []float64
	for rows2.Next() {
		var student_number float64
		err = rows2.Scan(&student_number)
		if err != nil {
			t.Errorf("rows2.Scan: %v", err)
		}
		sns = append(sns, student_number)
		fmt.Printf("student_number = %v\n", student_number)
	}
	if len(sns) != 2 {
		t.Errorf("len(sns): expected 2, got %v", len(sns))
	}

	err = rows.Err()
	if err != nil {
		t.Errorf("rows.Err: %v", err)
	}
}

func TestTeacherSched(t *testing.T) {
	config := GetConfig("ps.conf")
	db, err := sql.Open("oci8", MakeDSN(config))
	if err != nil {
		t.Fatal(err)
	}

	ch := GetTeacherSched(db, "fogel")
	var c int
	for mtg := range ch {
		if mtg.loginid != "fogel" {
			t.Errorf("mtg.loginid = %v, expected fogel")
		}
		c++
	}
	expected := 360
	if c != expected {
		t.Errorf("count is %v, expected %v", c, expected)
	}
}

func TestCalendar(t *testing.T) {
	config := GetConfig("ps.conf")
	db, err := sql.Open("oci8", MakeDSN(config))
	if err != nil {
		t.Fatal()
	}
	cal := GetCalendar(db)
	fmt.Print(cal)
}
