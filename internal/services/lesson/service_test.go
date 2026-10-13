package lesson

import (
	"errors"
	"strings"
	"testing"
	"time"
	"timetable-to-ics/internal/models"

	"github.com/xuri/excelize/v2"
)

func TestGetGroupsReturnsSortedUniqueGroupNames(t *testing.T) {
	workbook := excelize.NewFile()
	defer func() { _ = workbook.Close() }()

	for cell, value := range map[string]string{
		"A1": "Расписание", "C5": "ЦПИБВУ-31", "G5": "ЦИВТБв-11",
		"K5": "цпибву-31", "C6": "Предмет-12 с пробелом", "G6": "123-45",
	} {
		if err := workbook.SetCellValue("Sheet1", cell, value); err != nil {
			t.Fatal(err)
		}
	}

	data, err := workbook.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	groups, err := NewService().GetGroups([][]byte{data.Bytes()})
	if err != nil {
		t.Fatalf("GetGroups() error = %v", err)
	}

	want := []string{"ЦИВТБв-11", "ЦПИБВУ-31"}
	if len(groups) != len(want) {
		t.Fatalf("GetGroups() = %#v, want %#v", groups, want)
	}
	for i := range want {
		if groups[i] != want[i] {
			t.Errorf("GetGroups()[%d] = %q, want %q", i, groups[i], want[i])
		}
	}
}

func TestGetLessonSkipsRowsWithoutGroupColumn(t *testing.T) {
	service := &Service{loc: time.FixedZone("test", 0)}

	for _, currentColumns := range [][]string{
		nil,
		{},
		{"04.Sep"},
	} {
		_, err := service.getLesson(currentColumns, 71, 2026)
		if !errors.Is(err, EmptyErr) {
			t.Fatalf("getLesson(%v, 71) error = %v, want EmptyErr", currentColumns, err)
		}
	}
}

func TestAddOnlineLinks(t *testing.T) {
	service := &Service{}
	start := time.Date(2026, time.September, 15, 18, 0, 0, 0, time.UTC)
	lessons := []models.Lesson{
		{Name: "Экономика и управление проектами в IT-отрасли", StartTime: start},
		{Name: "Основы военной подготовки", StartTime: start.AddDate(0, 0, -4)},
	}
	links := []models.LessonLink{
		{Day: 15, Month: time.September, Subject: "Экономика и управление проектами в IT-отрасли", URL: "https://example.com/call"},
		{Day: 11, Month: time.September, Subject: "Основы военной подготовки - 1 курс", Info: "Подготовить презентацию и отправить преподавателю."},
	}

	got := service.AddOnlineLinks(lessons, links)
	if got[0].Link != "https://example.com/call" || got[0].Description != "Ссылка на онлайн-занятие: https://example.com/call" {
		t.Errorf("lesson with link = %#v", got[0])
	}
	if got[1].Link != "" || !strings.Contains(got[1].Description, "Подготовить презентацию") {
		t.Errorf("lesson without link = %#v", got[1])
	}
}

func TestAddOnlineLinksPrefersNewestAnnouncement(t *testing.T) {
	service := &Service{}
	lessons := []models.Lesson{{
		Name: "Системы ИИ", StartTime: time.Date(2026, time.September, 8, 18, 0, 0, 0, time.UTC),
	}}
	links := []models.LessonLink{
		{Day: 8, Month: time.September, Subject: "Системы ИИ"},
		{Day: 8, Month: time.September, Subject: "Системы ИИ", URL: "https://example.com/old"},
	}

	got := service.AddOnlineLinks(lessons, links)
	if got[0].Link != "" {
		t.Errorf("new empty announcement must override an old link, got %q", got[0].Link)
	}
}

func TestGetLessonsUsesYearFromWorkbook(t *testing.T) {
	workbook := excelize.NewFile()
	defer func() { _ = workbook.Close() }()

	for cell, value := range map[string]string{
		"A1": "Расписание занятий на 2027 год",
		"C5": "ЦПИБВУ-31",
		"B6": "04.Sep",
		"C6": "Databases",
	} {
		if err := workbook.SetCellValue("Sheet1", cell, value); err != nil {
			t.Fatal(err)
		}
	}
	data, err := workbook.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	lessons, err := NewService().GetLessons(models.GetCalendarRequest{Group: "цпибву-31"}, [][]byte{data.Bytes()})
	if err != nil {
		t.Fatalf("GetLessons() error = %v", err)
	}
	if len(lessons) != 1 {
		t.Fatalf("GetLessons() returned %d lessons, want 1", len(lessons))
	}
	if lessons[0].StartTime.Year() != 2027 {
		t.Errorf("lesson year = %d, want 2027", lessons[0].StartTime.Year())
	}
}

func TestGetLessonParsesLesson(t *testing.T) {
	service := &Service{loc: time.FixedZone("test", 0)}

	lesson, err := service.getLesson([]string{"04.Sep", "Databases"}, 1, 2026)
	if err != nil {
		t.Fatalf("getLesson() error = %v", err)
	}

	if lesson.Name != "Databases" {
		t.Errorf("lesson.Name = %q, want %q", lesson.Name, "Databases")
	}
	if want := time.Date(2026, time.September, 4, 18, 0, 0, 0, service.loc); !lesson.StartTime.Equal(want) {
		t.Errorf("lesson.StartTime = %v, want %v", lesson.StartTime, want)
	}
}

func TestParseLessonDateInfersClosestYearWhenWorkbookHasNoYear(t *testing.T) {
	service := &Service{
		loc: time.FixedZone("test", 0),
		now: func() time.Time {
			return time.Date(2026, time.December, 15, 12, 0, 0, 0, time.UTC)
		},
	}

	lessonDate, err := service.parseLessonDate("10.Jan", 0)
	if err != nil {
		t.Fatalf("parseLessonDate() error = %v", err)
	}
	if lessonDate.Year() != 2027 {
		t.Errorf("year = %d, want 2027", lessonDate.Year())
	}
}

func TestGetLessonParsesNumericDate(t *testing.T) {
	service := &Service{loc: time.FixedZone("test", 0)}

	lesson, err := service.getLesson([]string{"04.09", "Databases"}, 1, 2026)
	if err != nil {
		t.Fatalf("getLesson() error = %v", err)
	}
	if lesson.StartTime.Month() != time.September || lesson.StartTime.Day() != 4 {
		t.Errorf("lesson date = %v, want 4 September", lesson.StartTime)
	}
}
