package psfacade

import (
	"fmt"
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

	for rows2.Next() {
		var student_number float64
		err = rows2.Scan(&student_number)
		if err != nil {
			t.Errorf("rows2.Scan: %v", err)
		}
		fmt.Printf("student_number = %v\n", student_number)
	}
}
