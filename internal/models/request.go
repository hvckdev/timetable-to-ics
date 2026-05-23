package models

type GetCalendarRequest struct {
	Group    string `json:"group"`
	Timezone string `json:"timezone"`
}
