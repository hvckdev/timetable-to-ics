package groups

import (
	"encoding/json"
	"net/http"
	"timetable-to-ics/internal/models"
	groupsUsecase "timetable-to-ics/internal/usecase/groups"
)

type Handler struct {
	usecase *groupsUsecase.Usecase
}

func NewHandler(usecase *groupsUsecase.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	availableGroups, err := h.usecase.GetGroups(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, struct {
		Groups []string `json:"groups"`
	}{Groups: availableGroups})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
