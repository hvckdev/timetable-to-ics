package get_calendar

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"slices"
	"strings"
	"time"
	"timetable-to-ics/internal/clients/ulstu"
	"timetable-to-ics/internal/models"

	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type Usecase struct {
	ulstuClient *ulstu.Client
}

func NewUsecase(ulstuClient *ulstu.Client) *Usecase {
	return &Usecase{ulstuClient: ulstuClient}
}

func (uc *Usecase) GetCalendar(ctx context.Context, request models.GetCalendarRequest) ([]byte, error) {
	allFiles, err := uc.ulstuClient.ListLatestSchedules(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("get excels list: %w", err)
	}

	lessons := make([]models.Lesson, 0)
	for _, file := range allFiles {
		download, err := uc.ulstuClient.Download(ctx, file)
		if err != nil {
			return []byte{}, fmt.Errorf("excel download error: %w", err)
		}

		currentLessons, err := getLessons(request, download)
		if err != nil {
			return nil, err
		}

		lessons = append(lessons, currentLessons...)
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	for _, lesson := range lessons {
		event := cal.AddEvent(uuid.New().String())
		event.SetCreatedTime(time.Now())
		event.SetDtStampTime(time.Now())
		event.SetModifiedAt(time.Now())
		event.SetStartAt(lesson.StartTime)
		event.SetEndAt(lesson.EndTime)
		event.SetSummary(lesson.Name)
	}

	return []byte(cal.Serialize()), nil
}

func getLessons(request models.GetCalendarRequest, download []byte) ([]models.Lesson, error) {
	newReader := bytes.NewReader(download)
	reader, err := excelize.OpenReader(newReader)
	defer func(reader *excelize.File) {
		_ = reader.Close()
	}(reader)
	if err != nil {
		return []models.Lesson{}, fmt.Errorf("excel reader error: %w", err)
	}

	firstSheet := reader.GetSheetName(0)
	rows, err := reader.Rows(firstSheet)
	if err != nil {
		return nil, fmt.Errorf("get rows error: %w", err)
	}

	groupIndex := -1

	for rows.Next() && groupIndex < 1 {
		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("get columns error: %w", err)
		}

		groupIndex = slices.IndexFunc(cols, func(s string) bool {
			return strings.EqualFold(s, request.Group)
		})
	}

	lessons := make([]models.Lesson, 0)
	existsHash := make(map[string]bool)

	defer func(rows *excelize.Rows) {
		_ = rows.Close()
	}(rows)
	for rows.Next() {
		currentColumns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("get columns error: %w", err)
		}

		lessonInfo := currentColumns[groupIndex]
		lessonDate := currentColumns[groupIndex-1]

		if lessonInfo != "" {
			parse, err := time.Parse("02.Jan 2006 15:04 -07", lessonDate+" 2026 18:00 "+request.Timezone)
			if err != nil {
				return nil, fmt.Errorf("parse date error: %w", err)
			}

			hasher := md5.New()
			hash := hasher.Sum([]byte(lessonInfo + parse.String()))
			if existsHash[string(hash)] {
				continue
			}

			lessons = append(lessons, models.Lesson{
				Name:      lessonInfo,
				StartTime: parse,
				EndTime:   parse.Add(time.Hour * 3),
			})

			existsHash[string(hash)] = true
		}
	}

	return lessons, nil
}
