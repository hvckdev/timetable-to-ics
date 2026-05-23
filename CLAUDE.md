# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go CLI tool that converts ULSTU (Ulyanovsk State Technical University) Excel timetables into ICS calendar files. Built for students to avoid opening the .xlsx schedule daily.

## Build & Run

```bash
make all          # Build for all platforms (windows, linux, darwin)
make darwin       # Build for macOS (amd64, arm64)
make linux        # Build for linux (amd64, 386, arm, arm64)
make windows      # Build for windows (amd64, 386)
make clean        # Remove bin/
go build -o bin/timetable-to-ics ./cmd/   # Build locally
```

No tests exist yet. No CI/CD pipeline configured.

## Linting

GoLand uses golangci-lint with: `ineffassign`, `staticcheck`, `govet`, `errcheck`, `unused`, `gosimple`. No `.golangci.yml` file — config is in `.idea/golinter.xml`.

## Architecture

**All working logic currently lives in `cmd/main.go`.** The `internal/` packages are stubs from an incomplete refactor.

The CLI flow: prompt for .xlsx file path → prompt for group name → parse Excel rows for that group's column → create ICS events → write `result.ics`.

Key functions in `cmd/main.go`:
- `getColsFromExcelFile()` — opens Excel, reads all rows
- `getGroupName()` — prompts for group name (e.g. "ЦИСЭБву-31")
- `createCalendar()` — finds the group's column, iterates cells to build events
- `addOnlyInputGroupEventsToCalendar()` — parses dates, creates ICS events with fixed 18:00-20:00 timeslot
- `createResultAndSaveToFile()` — serializes and writes the ICS file

### In-Progress Refactor (uncommitted)

`internal/` contains stubs for a cleaner architecture:
- `internal/clients/ulstu-xlsx/` — intended XLSX parsing client
- `internal/handlers/` — intended HTTP handlers (WIP: `handleIndex` stub)
- `internal/models/lesson.go` — `Lesson` struct (Name, StartTime, EndTime, Link)
- `internal/usecase/` — intended ICS generation business logic

`internal/infrastructure/export/` was deleted in staging — had a Strategy interface that was removed.

### Hardcoded Configuration

- Timezone: `Europe/Samara` (UTC+4)
- Lesson times: fixed 18:00-20:00 (evening classes)
- Date format: `"2/Jan 2006 15:04"` (Russian-style dates from Excel)
- Output: `result.ics` in working directory
- Year: assumes `time.Now().Year()` for all dates

## Key Dependencies

- `github.com/xuri/excelize/v2` — Excel file parsing
- `github.com/arran4/golang-ical` — ICS calendar generation
- `github.com/google/uuid` — unique event IDs
- `time/tzdata` — embedded timezone DB (Windows compatibility)

## Important Details

- The Excel parsing assumes ULSTU-specific row layout: dates are in the row **above** the lesson name, room numbers are in the row **below**
- Event descriptions are in Russian (`"Аудитория"` = "Room")
- `ics.Calendar` is **not thread-safe** — a prior attempt at concurrency caused issues (see git history: commit `b1c62f5` added `go`, `fd970d6` removed it)
- The working tree has WIP changes adding an HTTP server on `:8080` — currently blocks on `ListenAndServe`, making CLI logic unreachable