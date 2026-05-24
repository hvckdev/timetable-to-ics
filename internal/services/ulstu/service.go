package ulstu

import (
	"context"
	"fmt"
	"timetable-to-ics/internal/clients/ulstu"
)

type Service struct {
	ulstuClient *ulstu.Client
}

func NewService(ulstuClient *ulstu.Client) *Service {
	return &Service{ulstuClient: ulstuClient}
}

func (s *Service) GetAllFilesData(ctx context.Context) ([][]byte, error) {
	filesInfo, err := s.ulstuClient.ListLatestSchedules(ctx)
	if err != nil {
		return [][]byte{}, fmt.Errorf("get excels list: %w", err)
	}

	allFiles := make([][]byte, 0)
	for _, fileInfo := range filesInfo {
		download, err := s.ulstuClient.Download(ctx, fileInfo)
		if err != nil {
			return [][]byte{}, fmt.Errorf("excel download error: %w", err)
		}

		allFiles = append(allFiles, download)
	}

	return allFiles, nil
}
