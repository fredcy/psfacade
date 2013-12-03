package main

import (
	"fmt"
	_ "github.com/fredcy/icalendar"
	"github.com/fredcy/psfacade"
	_ "log"
)

func main() {
	cal := psfacade.TeacherCalendar("fogel")
	fmt.Print(cal)
}
