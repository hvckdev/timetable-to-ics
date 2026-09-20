package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	_ "time/tzdata"
	ulstu_xlsx "timetable-to-ics/internal/clients/ulstu"
	get_calendar "timetable-to-ics/internal/handlers/get-calendar"
	groupsHandler "timetable-to-ics/internal/handlers/groups"
	"timetable-to-ics/internal/handlers/index"
	"timetable-to-ics/internal/handlers/web"
	"timetable-to-ics/internal/services/calendar"
	"timetable-to-ics/internal/services/lesson"
	"timetable-to-ics/internal/services/ulstu"
	get_calendar2 "timetable-to-ics/internal/usecase/get-calendar"
	groupsUsecase "timetable-to-ics/internal/usecase/groups"
	index2 "timetable-to-ics/internal/usecase/index"
)

func main() {
	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8589"
	}

	// clients
	ulstuClient := ulstu_xlsx.NewClient()

	// services
	ulstuService := ulstu.NewService(ulstuClient)
	lessonService := lesson.NewService()
	calendarService := calendar.NewService()

	// use cases
	indexUsecase := index2.NewUsecase(ulstuClient)
	getCalendarUsecase := get_calendar2.NewUsecase(lessonService, ulstuService, calendarService)
	getGroupsUsecase := groupsUsecase.NewUsecase(lessonService, ulstuService)

	// handlers
	indexHandler := index.NewIndexHandler(ctx, indexUsecase)
	getCalendarHandler := get_calendar.NewHandler(ctx, getCalendarUsecase)
	getGroupsHandler := groupsHandler.NewHandler(getGroupsUsecase)
	webHandler := web.NewHandler()

	http.Handle("/api/schedules", indexHandler)
	http.Handle("/api/groups", getGroupsHandler)
	http.Handle("/calendar", getCalendarHandler)
	http.Handle("/assets/", http.StripPrefix("/assets/", webHandler.Assets()))
	http.Handle("/", webHandler)

	addr := ":" + port
	fmt.Printf("Server starting on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
