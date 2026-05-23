package index

import (
	"context"
	"encoding/json"
	"net/http"
	"timetable-to-ics/internal/models"
	"timetable-to-ics/internal/usecase/index"
)

type Handler struct {
	usecase *index.Usecase
	ctx     context.Context
}

func NewIndexHandler(ctx context.Context, usecase *index.Usecase) *Handler {
	return &Handler{ctx: ctx, usecase: usecase}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	result, err := h.usecase.GetAllFiles(h.ctx)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Error: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		marshal, _ := json.Marshal(errorResponse)
		_, _ = w.Write(marshal)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}
