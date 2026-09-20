package lesson

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"timetable-to-ics/internal/models"
	"unicode"

	"github.com/xuri/excelize/v2"
)

var EmptyErr = fmt.Errorf("lesson is empty")

const onlineLessonsPage = "https://coe.ulstu.ru/index.php?action=show_page&id=103"

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

// AddOnlineLinks enriches lessons with announcements from the ULSTU online lessons page.
func (s *Service) AddOnlineLinks(lessons []models.Lesson, links []models.LessonLink) []models.Lesson {
	type lessonKey struct {
		day     int
		month   time.Month
		subject string
	}

	linksByLesson := make(map[lessonKey]models.LessonLink, len(links))
	for _, link := range links {
		key := lessonKey{day: link.Day, month: link.Month, subject: normalizeLessonName(link.Subject)}
		if _, exists := linksByLesson[key]; !exists {
			// The newest announcements are at the top of the page, so the first one wins.
			linksByLesson[key] = link
		}
	}

	for i := range lessons {
		key := lessonKey{
			day: lessons[i].StartTime.Day(), month: lessons[i].StartTime.Month(),
			subject: normalizeLessonName(lessons[i].Name),
		}
		announcement, exists := linksByLesson[key]
		if exists && announcement.URL != "" {
			lessons[i].Link = announcement.URL
			lessons[i].Description = "Ссылка на онлайн-занятие: " + announcement.URL
			continue
		}
		if exists && announcement.Info != "" {
			lessons[i].Description = "Информация со страницы онлайн-занятий:\n" + announcement.Info + "\n\nИсточник: " + onlineLessonsPage
			continue
		}
		lessons[i].Description = "Ссылка на онлайн-занятие пока не опубликована. Проверяйте страницу: " + onlineLessonsPage
	}

	return lessons
}

func normalizeLessonName(value string) string {
	value = strings.ToLower(strings.ReplaceAll(value, "ё", "е"))
	value = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return ' '
	}, value)

	fields := strings.Fields(value)
	if len(fields) >= 2 && fields[len(fields)-1] == "курс" && isDigits(fields[len(fields)-2]) {
		fields = fields[:len(fields)-2]
	}
	return strings.Join(fields, " ")
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (s *Service) getLessonFromFileData(fileData []byte, request models.GetCalendarRequest) ([]models.Lesson, error) {
	lessons := make([]models.Lesson, 0)

	newReader := bytes.NewReader(fileData)
	reader, err := excelize.OpenReader(newReader)
	if err != nil {
		return nil, fmt.Errorf("excel reader error: %w", err)
	}
	defer func() { _ = reader.Close() }()

	firstSheet := reader.GetSheetName(0)
	if firstSheet == "" {
		return nil, fmt.Errorf("get first sheet: workbook has no sheets")
	}

	rows, err := reader.Rows(firstSheet)
	if err != nil {
		return nil, fmt.Errorf("get rows error: %w", err)
	}
	defer func() { _ = rows.Close() }()

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
		if errors.Is(err, EmptyErr) {
			continue
		}
		if err != nil {
			return nil, err
		}

		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func (s *Service) getLesson(currentColumns []string, groupIndex int) (models.Lesson, error) {
	// Excelize omits trailing empty cells, so rows outside a group's block can be shorter.
	if groupIndex <= 0 || groupIndex >= len(currentColumns) {
		return models.Lesson{}, EmptyErr
	}

	lessonInfo := strings.TrimSpace(currentColumns[groupIndex])
	if lessonInfo == "" {
		return models.Lesson{}, EmptyErr
	}

	lessonDate := strings.TrimSpace(currentColumns[groupIndex-1])
	if lessonDate == "" {
		return models.Lesson{}, fmt.Errorf("lesson date is empty for %q", lessonInfo)
	}

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
