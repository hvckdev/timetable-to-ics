package lesson

import (
	"errors"
	"testing"
	"time"
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
