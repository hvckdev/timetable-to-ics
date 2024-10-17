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
	_ "time/tzdata"
)

const layout = "2/Jan 2006 15:04"
const resultFileName = "result.ics"
const lessonStartAt = "18:00"
const lessonEndAt = "20:00"
const currentLocation = "Europe/Samara"

func main() {
	fmt.Println("------")
	fmt.Printf("coo ulstu timetable to ics by %s", termlink.Link("hvck", "https://hvck.dev"))
	fmt.Println()
	fmt.Println("------")

	cols := getColsFromExcelFile()

	groupName := getGroupName()

	loc := getLocation()

	cal := createCalendar(cols, groupName, loc)

	defer func() {
		createResultAndSaveToFile(cal)
		fmt.Println("done")

		fmt.Println("click 'enter' for exit...")
		_, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
		if err != nil {
			log.Fatalf(err.Error())
		}
	}()
}

func getColsFromExcelFile() [][]string {
	var filePath string

	fmt.Print("enter absolute filepath to timetable >> ")

	_, err := fmt.Scan(&filePath)
	if err != nil {
		log.Fatalf("Failed to get filepath")
	}

	file, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("failed to open Excel file")
	}

	cols, err := file.GetRows(file.GetSheetName(0))
	if err != nil {
		log.Fatalf("Failed to get cols")
	}

	return cols
}

func getGroupName() string {
	var groupName string

	fmt.Println()
	fmt.Print("enter your group name >> ")

	_, err := fmt.Scan(&groupName)
	if err != nil {
		log.Fatalf("failed to receive group name")
	}
	return groupName
}

func getLocation() *time.Location {
	fmt.Println("trying to get current location")
	loc, err := time.LoadLocation(currentLocation)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Error loading location"))
	}
	fmt.Println("successfully got location")
	return loc
}

func createCalendar(cols [][]string, groupName string, loc *time.Location) *ics.Calendar {
	fmt.Println("creating calendar")

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	var groupRowIndex int

	fmt.Println("finding your group")

	for colIndex, row := range cols {
		for rowIndex, colCell := range row {
			if colCell == groupName {
				groupRowIndex = rowIndex
				fmt.Println("group found")
			}

			addOnlyInputGroupEventsToCalendar(cols, groupRowIndex, rowIndex, colIndex, colCell, loc, cal)
		}
	}

	return cal
}

func addOnlyInputGroupEventsToCalendar(cols [][]string, groupRowIndex int, rowIndex int, colIndex int, colCell string, loc *time.Location, cal *ics.Calendar) {
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

func getParsedDate(date string, timeString string, loc *time.Location) time.Time {
	t, err := time.ParseInLocation(layout, fmt.Sprintf("%s %d %s", date, time.Now().Year(), timeString), loc)
	if err != nil {
		log.Panicf(fmt.Sprintf("Error parsing date: %s", err))
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
		log.Panicf("failed to create file: %v", err)
	}
	defer func(result *os.File) {
		err := result.Close()
		if err != nil {
			log.Panicf("failed to write data to file: %v", err)
		}
	}(result)

	writer := bufio.NewWriter(result)

	_, err = writer.Write([]byte(cal.Serialize()))
	if err != nil {
		log.Panicf("failed to write data to file: %v", err)
	}

	if err := writer.Flush(); err != nil {
		log.Panicf("failed to flush writer: %v", err)
	}
}
