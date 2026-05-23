package models

import "time"

type File struct {
	DisplayName string     `json:"display_name"`
	URL         string     `json:"url"`
	Month       time.Month `json:"month"`
	Year        int        `json:"year"`
}
