package psfacade

import (
	"time"
)

func get_yearid() int {
	now := time.Now()
	var academicyear int
	if now.Month() < 7 {
		academicyear = now.Year()
	} else {
		academicyear = now.Year() + 1
	}
	yearid := academicyear - 1991 // the usual PowerSchool conversion
	return yearid
}
