package psfacade

import (
	"log"
	"database/sql"
)

type Student struct {
	Number string
	First_name string
	Last_name string
	Room string
	Birthdate string
}

var student_query = `
select to_char(student_number) "student_number",
first_name, last_name,
ps_customfields.getStudentscf(id, 'IMSA_Student_Room') room,
to_char(dob, 'YYYYMMDD') dob
from students where schoolid = 140177
and enroll_status = 0
order by last_name, first_name
`

func GetStudents(db *sql.DB) <-chan Student {
	students := make(chan Student)
	rows, err := db.Query(student_query)
	if err != nil {
		log.Printf("ERROR: query failed: %v", err)
		log.Printf("query=\"%v\"", student_query)
		close(students)
		return students
	}
	go func() {
		defer close(students)
		defer rows.Close()
		for rows.Next() {
			student := Student{}
			var dob string
			err := rows.Scan(&student.Number, &student.First_name, &student.Last_name, &student.Room, &dob)
			student.Birthdate = dob
			if err != nil {
				log.Printf("rows.Scan: ", err)
				return
			}
			students <- student
		}
		err := rows.Err()
		if err != nil {
			log.Printf("ERROR: %v", err)
		}
	}()
	return students
}
