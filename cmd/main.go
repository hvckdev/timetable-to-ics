package main

import (
	"bufio"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/savioxavier/termlink"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"time"
)

const layout = "2-Jan 2006 15:04"
const resultFileName = "result.ics"
const lessonStartAt = "18:00"
const lessonEndAt = "21:00"
const currentLocation = "Europe/Samara"

func main() {
	fmt.Println("------")
	fmt.Println(fmt.Sprintf("coo ulstu timetable to ics by %s", termlink.Link("hvck", "https://hvck.dev")))
	fmt.Println("------")

	cols := getColsFromExcelFile()

	groupName := getGroupName()

	loc := getLocation()

	cal := createCalendar(cols, groupName, loc)

	createResultAndSaveToFile(cal)
}

func getColsFromExcelFile() [][]string {
	var filePath string

	fmt.Print("enter absolute filepath to timetable >> ")

	_, err := fmt.Scan(&filePath)
	if err != nil {
		panic("Failed to get filepath")
	}

	file, err := excelize.OpenFile(filePath)
	if err != nil {
		panic("failed to open Excel file")
	}

	cols, err := file.GetRows(file.GetSheetName(0))
	if err != nil {
		panic("Failed to get cols")
	}

	return cols
}

func getGroupName() string {
	var groupName string

	fmt.Println()
	fmt.Print("enter your group name >> ")

	_, err := fmt.Scan(&groupName)
	if err != nil {
		panic("failed to receive group name")
	}
	return groupName
}

func getLocation() *time.Location {
	loc, err := time.LoadLocation(currentLocation)
	if err != nil {
		panic(fmt.Sprintf("Error loading location: %s", err))
	}
	return loc
}

func createCalendar(cols [][]string, groupName string, loc *time.Location) *ics.Calendar {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	var groupRowIndex int

	for colIndex, row := range cols {
		for rowIndex, colCell := range row {
			if colCell == groupName {
				groupRowIndex = rowIndex
			}

			if groupRowIndex == rowIndex {
				if rowIndex > 1 {
					date := cols[colIndex][rowIndex-1]

					if len(date) > 3 && len(colCell) > 0 {
						lessonStartDate := getParsedDate(date, lessonStartAt, loc)
						lessonEndDate := getParsedDate(date, lessonEndAt, loc)
						room := cols[colIndex][rowIndex+1]

						addEventToCalendar(cal, lessonStartDate, lessonEndDate, room, colCell)
					}
				}
			}
		}
	}

	return cal
}

func getParsedDate(date string, timeString string, loc *time.Location) time.Time {
	t, err := time.ParseInLocation(layout, fmt.Sprintf("%s %d %s", date, time.Now().Year(), timeString), loc)
	if err != nil {
		panic(fmt.Sprintf("Error parsing date: %s", err))
	}
	return t
}

func addEventToCalendar(cal *ics.Calendar, start time.Time, end time.Time, room string, colCell string) {
	event := cal.AddEvent(uuid.New().String())
	event.SetCreatedTime(time.Now())
	event.SetDtStampTime(time.Now())
	event.SetModifiedAt(time.Now())
	event.SetStartAt(start)
	event.SetEndAt(end)
	event.SetDescription(fmt.Sprintf("Аудитория %s", room))
	event.SetSummary(colCell)
}

func createResultAndSaveToFile(cal *ics.Calendar) {
	result, err := os.Create(resultFileName)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer result.Close()

	writer := bufio.NewWriter(result)

	_, err = writer.Write([]byte(cal.Serialize()))
	if err != nil {
		log.Fatalf("failed to write data to file: %v", err)
	}

	if err := writer.Flush(); err != nil {
		log.Fatalf("failed to flush writer: %v", err)
	}
}
