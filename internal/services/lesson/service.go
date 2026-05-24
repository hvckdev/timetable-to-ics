package lesson

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"timetable-to-ics/internal/models"

	"github.com/xuri/excelize/v2"
)

var EmptyErr = fmt.Errorf("lesson is empty")

type Service struct {
	loc *time.Location
}

func NewService() *Service {
	location, err := time.LoadLocation("Europe/Ulyanovsk")
	if err != nil {
		panic(err)
	}

	return &Service{loc: location}
}

func (s *Service) GetLessons(request models.GetCalendarRequest, filesData [][]byte) ([]models.Lesson, error) {
	result := make([]models.Lesson, 0)

	for _, fileData := range filesData {
		lessons, err := s.getLessonFromFileData(fileData, request)
		if err != nil {
			return nil, fmt.Errorf("get lessons: %w", err)
		}

		result = append(result, lessons...)
	}

	return result, nil
}

func (s *Service) getLessonFromFileData(fileData []byte, request models.GetCalendarRequest) ([]models.Lesson, error) {
	lessons := make([]models.Lesson, 0)

	newReader := bytes.NewReader(fileData)
	reader, err := excelize.OpenReader(newReader)
	defer func(reader *excelize.File) {
		_ = reader.Close()
	}(reader)
	if err != nil {
		return nil, fmt.Errorf("excel reader error: %w", err)
	}

	firstSheet := reader.GetSheetName(0)
	rows, err := reader.Rows(firstSheet)
	defer func(rows *excelize.Rows) {
		_ = rows.Close()
	}(rows)
	if err != nil {
		return nil, fmt.Errorf("get rows error: %w", err)
	}

	groupIndex, err := getGroupIndex(rows, request)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		currentColumns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("get columns error: %w", err)
		}

		lesson, err := s.getLesson(currentColumns, groupIndex)
		if err != nil && !errors.Is(err, EmptyErr) {
			return nil, err
		}

		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func (s *Service) getLesson(currentColumns []string, groupIndex int) (models.Lesson, error) {
	lessonInfo := currentColumns[groupIndex]
	if lessonInfo == "" {
		return models.Lesson{}, EmptyErr
	}

	lessonDate := currentColumns[groupIndex-1]

	parse, err := time.ParseInLocation("02.Jan 2006 15:04", lessonDate+" 2026 18:00", s.loc)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("parse date error: %w", err)
	}

	lesson := models.Lesson{
		Name:      lessonInfo,
		StartTime: parse,
		EndTime:   parse.Add(time.Hour * 3),
	}
	return lesson, nil
}

func getGroupIndex(rows *excelize.Rows, request models.GetCalendarRequest) (int, error) {
	groupIndex := -1

	for rows.Next() && groupIndex < 1 {
		cols, err := rows.Columns()
		if err != nil {
			return -1, fmt.Errorf("get columns error: %w", err)
		}

		groupIndex = slices.IndexFunc(cols, func(s string) bool {
			return strings.EqualFold(s, request.Group)
		})
	}

	if groupIndex == -1 {
		return -1, fmt.Errorf("get group index error: %w", EmptyErr)
	}

	return groupIndex, nil
}
