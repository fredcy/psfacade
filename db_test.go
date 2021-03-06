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
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	query2 := "SELECT student_number from Students where last_name = :1"
	rows2, err2 := RunQuery(config, query2, "Yan")
	if err2 != nil {
		t.Fatalf("query2 failed: %v", err2)
	}
	defer rows2.Close()

	var sns []float64
	for rows2.Next() {
		var studentNumber float64
		err = rows2.Scan(&studentNumber)
		if err != nil {
			t.Errorf("rows2.Scan: %v", err)
		}
		sns = append(sns, studentNumber)
		fmt.Printf("studentNumber = %v\n", studentNumber)
	}
	var expected int = 3
	if len(sns) != expected {
		t.Errorf("len(sns): expected %v, got %v", expected, len(sns))
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
			t.Errorf("mtg.loginid = %v, expected fogel", mtg.loginid)
		}
		c++
	}
	if c < 100 || c > 500 {
		t.Errorf("count is %v, expected in range [100..500]", c)
	}
}

func TestCalendar(t *testing.T) {
	config := GetConfig("ps.conf")
	db, err := sql.Open("oci8", MakeDSN(config))
	if err != nil {
		t.Fatal(err)
	}
	calstr := GetCalendar(db).String()
	callen := len(calstr)
	if testing.Verbose() {
		fmt.Print(calstr, callen)
	}
	if callen < 20000 || callen > 500000 {
		t.Errorf("generated calendar length (%v) is not valid", callen)
	}
}
