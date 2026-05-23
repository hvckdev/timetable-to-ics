package index

import (
	ulstuxlsx "timetable-to-ics/internal/clients/ulstu"
	"timetable-to-ics/internal/models"
)

func convertFiles(allFiles []ulstuxlsx.ScheduleFile) []models.File {
	convertedFiles := make([]models.File, 0, len(allFiles))
	for _, file := range allFiles {
		newFile := models.File{
			URL:         file.URL,
			DisplayName: file.DisplayName,
			Month:       file.Month,
			Year:        file.Year,
		}

		convertedFiles = append(convertedFiles, newFile)
	}
	return convertedFiles
}
