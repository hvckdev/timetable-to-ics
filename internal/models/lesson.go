package models

import "time"

type Lesson struct {
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Link      string    `json:"link"`
}
