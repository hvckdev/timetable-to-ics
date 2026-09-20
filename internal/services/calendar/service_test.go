package calendar

import (
	"strings"
	"testing"
	"time"

	"timetable-to-ics/internal/models"
)

func TestMakeCalendarAddsOnlineLessonDetails(t *testing.T) {
	lesson := models.Lesson{
		Name:        "Системы ИИ",
		StartTime:   time.Date(2026, time.September, 8, 18, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2026, time.September, 8, 21, 0, 0, 0, time.UTC),
		Link:        "https://example.com/call",
		Description: "Ссылка на онлайн-занятие: https://example.com/call",
	}

	serialized := NewService().MakeCalendarFromLessons([]models.Lesson{lesson}).Serialize()
	if !strings.Contains(serialized, "URL:https://example.com/call") {
		t.Errorf("calendar has no URL property:\n%s", serialized)
	}
	if !strings.Contains(serialized, "DESCRIPTION:Ссылка на онлайн-занятие:") {
		t.Errorf("calendar has no description:\n%s", serialized)
	}
}

func TestMakeCalendarAddsLessonReminders(t *testing.T) {
	lesson := models.Lesson{
		Name:      "Системы ИИ",
		StartTime: time.Date(2026, time.September, 8, 18, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, time.September, 8, 21, 0, 0, 0, time.UTC),
	}

	serialized := NewService().MakeCalendarFromLessons([]models.Lesson{lesson}).Serialize()
	if count := strings.Count(serialized, "BEGIN:VALARM"); count != 2 {
		t.Errorf("calendar has %d reminders instead of 2:\n%s", count, serialized)
	}
	for _, expected := range []string{
		"ACTION:DISPLAY",
		"TRIGGER:-PT1H",
		"TRIGGER:-PT3M",
		"DESCRIPTION:Скоро начнётся занятие: Системы ИИ",
	} {
		if !strings.Contains(serialized, expected) {
			t.Errorf("calendar has no %q:\n%s", expected, serialized)
		}
	}
}
