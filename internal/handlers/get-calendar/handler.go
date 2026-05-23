package get_calendar

import (
	"context"
	"encoding/json"
	"net/http"
	"timetable-to-ics/internal/models"
	get_calendar "timetable-to-ics/internal/usecase/get-calendar"
)

type GetCalendarHandler struct {
	usecase *get_calendar.Usecase
	context context.Context
}

func NewHandler(context context.Context, usecase *get_calendar.Usecase) *GetCalendarHandler {
	return &GetCalendarHandler{context: context, usecase: usecase}
}

func (h *GetCalendarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	calendar, err := h.usecase.GetCalendar(h.context, models.GetCalendarRequest{Group: r.URL.Query().Get("group")})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := models.ErrorResponse{Error: err.Error()}
		marshal, _ := json.Marshal(response)
		_, _ = w.Write(marshal)
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"calendar.ics\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(calendar)
}
