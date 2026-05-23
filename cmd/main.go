package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	_ "time/tzdata"
	ulstu_xlsx "timetable-to-ics/internal/clients/ulstu"
	get_calendar "timetable-to-ics/internal/handlers/get-calendar"
	"timetable-to-ics/internal/handlers/index"
	get_calendar2 "timetable-to-ics/internal/usecase/get-calendar"
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

	// use cases
	indexUsecase := index2.NewUsecase(ulstuClient)
	getCalendarUsecase := get_calendar2.NewUsecase(ulstuClient)

	// handlers
	indexHandler := index.NewIndexHandler(ctx, indexUsecase)
	getCalendarHandler := get_calendar.NewHandler(ctx, getCalendarUsecase)

	http.Handle("/", indexHandler)
	http.Handle("/calendar", getCalendarHandler)

	addr := ":" + port
	fmt.Printf("Server starting on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
