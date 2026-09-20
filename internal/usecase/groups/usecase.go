package groups

import (
	"context"
	"fmt"
	"sync"
	"time"
	"timetable-to-ics/internal/services/lesson"
	"timetable-to-ics/internal/services/ulstu"
)

const cacheTTL = 15 * time.Minute

type Usecase struct {
	lessonService *lesson.Service
	ulstuService  *ulstu.Service

	mu        sync.Mutex
	groups    []string
	expiresAt time.Time
}

func NewUsecase(lessonService *lesson.Service, ulstuService *ulstu.Service) *Usecase {
	return &Usecase{lessonService: lessonService, ulstuService: ulstuService}
}

func (uc *Usecase) GetGroups(ctx context.Context) ([]string, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if time.Now().Before(uc.expiresAt) {
		return append([]string(nil), uc.groups...), nil
	}

	filesData, err := uc.ulstuService.GetAllFilesData(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all files data: %w", err)
	}

	availableGroups, err := uc.lessonService.GetGroups(filesData)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}

	uc.groups = append([]string(nil), availableGroups...)
	uc.expiresAt = time.Now().Add(cacheTTL)

	return append([]string(nil), availableGroups...), nil
}
