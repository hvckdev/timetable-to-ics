package calendar

import (
	"crypto/md5"
	"time"

	"timetable-to-ics/internal/models"

	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) MakeCalendarFromLessons(lessons []models.Lesson) *ics.Calendar {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	hasher := md5.New()
	existsHash := make(map[string]bool)
	for _, lesson := range lessons {
		hash := hasher.Sum([]byte(lesson.Name + lesson.StartTime.String()))
		if existsHash[string(hash)] {
			continue
		}

		event := cal.AddEvent(uuid.New().String())
		event.SetCreatedTime(time.Now())
		event.SetDtStampTime(time.Now())
		event.SetModifiedAt(time.Now())
		event.SetStartAt(lesson.StartTime)
		event.SetEndAt(lesson.EndTime)
		event.SetSummary(lesson.Name)
		event.SetDescription(lesson.Description)
		if lesson.Link != "" {
			event.SetURL(lesson.Link)
		}

		addReminder(event, lesson.Name, "-PT1H")
		addReminder(event, lesson.Name, "-PT3M")

		existsHash[string(hash)] = true
	}

	return cal
}

func addReminder(event *ics.VEvent, lessonName, trigger string) {
	alarm := event.AddAlarm()
	alarm.SetAction(ics.ActionDisplay)
	alarm.SetDescription("Скоро начнётся занятие: " + lessonName)
	alarm.SetTrigger(trigger)
}
