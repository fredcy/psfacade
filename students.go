package psfacade

import (
	"database/sql"
	"log"
)

// Student is the student data obtained from PowerSchool
type Student struct {
	Number    string
	FirstName string
	LastName  string
	Room      string
	Username  string
}

var studentQuery = `
select to_char(student_number) "student_number",
first_name, last_name,
ps_customfields.getStudentscf(id, 'IMSA_Student_Room') room,
--to_char(dob, 'YYYY-MM-DD') dob,
student_web_id username
from students where schoolid = 140177
and enroll_status = 0
order by last_name, first_name
`

// GetStudents reads the PowerSchool database and returns a channel of Student values
func GetStudents(db *sql.DB) <-chan Student {
	students := make(chan Student)
	rows, err := db.Query(studentQuery)
	if err != nil {
		log.Printf("ERROR: query failed: %v", err)
		log.Printf("query=\"%v\"", studentQuery)
		close(students)
		return students
	}
	go func() {
		defer close(students)
		defer rows.Close()
		for rows.Next() {
			student := Student{}
			err := rows.Scan(&student.Number, &student.FirstName, &student.LastName, &student.Room,
				&student.Username)
			if err != nil {
				log.Printf("rows.Scan: %v", err)
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
