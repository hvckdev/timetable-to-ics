package lesson

import (
	"errors"
	"strings"
	"testing"
	"time"
	"timetable-to-ics/internal/models"
)

func TestGetLessonSkipsRowsWithoutGroupColumn(t *testing.T) {
	service := &Service{loc: time.FixedZone("test", 0)}

	for _, currentColumns := range [][]string{
		nil,
		{},
		{"04.Sep"},
	} {
		_, err := service.getLesson(currentColumns, 71)
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

func TestGetLessonParsesLesson(t *testing.T) {
	service := &Service{loc: time.FixedZone("test", 0)}

	lesson, err := service.getLesson([]string{"04.Sep", "Databases"}, 1)
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
