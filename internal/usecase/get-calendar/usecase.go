package get_calendar

import (
	"context"
	"fmt"
	"timetable-to-ics/internal/models"
	"timetable-to-ics/internal/services/calendar"
	"timetable-to-ics/internal/services/lesson"
	"timetable-to-ics/internal/services/ulstu"
)

type Usecase struct {
	lessonService   *lesson.Service
	ulstuService    *ulstu.Service
	calendarService *calendar.Service
}

func NewUsecase(
	lessonService *lesson.Service,
	ulstuService *ulstu.Service,
	calendarService *calendar.Service,
) *Usecase {
	return &Usecase{
		lessonService:   lessonService,
		ulstuService:    ulstuService,
		calendarService: calendarService,
	}
}

func (uc *Usecase) GetCalendar(ctx context.Context, request models.GetCalendarRequest) ([]byte, error) {
	filesData, err := uc.ulstuService.GetAllFilesData(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all files data: %w", err)
	}

	lessons, err := uc.lessonService.GetLessons(request, filesData)
	if err != nil {
		return nil, fmt.Errorf("get lessons: %w", err)
	}

	cal := uc.calendarService.MakeCalendarFromLessons(lessons)

	return []byte(cal.Serialize()), nil
}
