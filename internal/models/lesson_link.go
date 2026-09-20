package models

import "time"

// LessonLink describes an online lesson announcement published on the ULSTU page.
type LessonLink struct {
	Day     int
	Month   time.Month
	Subject string
	URL     string
	Info    string
}
