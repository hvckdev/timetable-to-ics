package index

import (
	"context"
	"encoding/json"
	"fmt"
	ulstuxlsx "timetable-to-ics/internal/clients/ulstu"
)

type Usecase struct {
	ulstuClient *ulstuxlsx.Client
}

func NewUsecase(ulstuClient *ulstuxlsx.Client) *Usecase {
	return &Usecase{ulstuClient: ulstuClient}
}

func (uc *Usecase) GetAllFiles(ctx context.Context) ([]byte, error) {
	allFiles, err := uc.ulstuClient.ListLatestSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListLatestSchedules: %w", err)
	}

	convertedFiles := convertFiles(allFiles)

	return json.Marshal(convertedFiles)
}
