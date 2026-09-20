# CLAUDE.md

This file provides guidance for coding agents working in this repository.

## Project Overview

`timetable-to-ics` is a Go HTTP service that converts ULSTU (Ulyanovsk State Technical University) Excel timetables into downloadable ICS calendars.

The service:

1. Reads the ULSTU Center for Online Education page:
   `https://coe.ulstu.ru/index.php?action=show_page&id=103`.
2. Finds and downloads the latest attached `.xlsx` timetable for each month.
3. Extracts lessons for a requested student group.
4. Parses online lesson links and announcements from the same page.
5. Matches announcements to lessons by date and normalized subject name.
6. Returns an ICS calendar containing lesson times, links, and descriptions.

This is now an HTTP service, not the original interactive CLI described in older revisions.

## Build, Run, and Test

```bash
# Run locally on the default port 8589
go run ./cmd

# Run on another port
PORT=8080 go run ./cmd

# Build locally
go build -o bin/timetable-to-ics ./cmd

# Run all tests
go test ./...

# Build all configured platform binaries
make all
make darwin
make linux
make windows
make clean

# Run with Docker Compose
docker compose up --build
```

There is currently no CI/CD pipeline.

## HTTP API

### `GET /`

Returns JSON metadata for the latest discovered timetable files, including display name, URL, month, and year.

### `GET /calendar?group=<group>`

Returns a generated calendar as `text/calendar; charset=utf-8` with filename `calendar.ics`.

Example:

```text
http://localhost:8589/calendar?group=цпибву-31
```

A missing `group` parameter returns HTTP 422. Timetable/parsing failures currently return HTTP 400 from the calendar handler.

## Architecture

Application wiring is in `cmd/main.go`:

- Creates the ULSTU client.
- Creates services and use cases.
- Registers `/` and `/calendar` handlers.
- Starts `net/http` on `PORT`, defaulting to `8589`.

### `internal/clients/ulstu/`

External ULSTU integration:

- Fetches the online education page using Windows-1251 decoding.
- Finds timetable attachments under the `Приложения:` section.
- Selects the last uploaded schedule for each month.
- Downloads `.xlsx` files.
- Parses the page's free-form online lesson announcements into `models.LessonLink` values.

The announcements are not a table. They are blocks separated by long underscore lines. The parser supports:

- `Предмет:` and `Предметы:`.
- One or several dates in a block.
- One or several slash-separated subjects.
- URLs in visible text or only in an anchor `href`.
- URLs placed directly after the `Ссылка:` label.
- Empty URL announcements containing only instructions or other information.
- Duplicate page content in the HTML DOM.

The page lists newest announcements first. Deduplication and matching intentionally preserve the first occurrence so a newer empty announcement can supersede an older link.

### `internal/services/ulstu/`

Coordinates timetable downloads and loading online lesson announcements through the client.

### `internal/services/lesson/`

Parses downloaded Excel workbooks and enriches lessons with online announcements.

Important behavior:

- Only the first workbook sheet is read.
- A group is found by case-insensitive exact comparison with a cell value.
- The lesson name is read from the group's column.
- The date is expected in the immediately preceding column of the same row.
- Empty/truncated Excel rows are skipped because Excelize omits trailing empty cells.
- Announcement matching uses day, month, and a normalized subject name.
- Normalization ignores case, punctuation, `ё`/`е`, repeated whitespace, and a trailing marker such as `- 1 курс`.
- If an announcement has a URL, it becomes the lesson link and description.
- If it has no URL, its published text is copied into the event description with the source page URL.
- If no useful announcement exists, the description says that the link is not published and points to the source page.

### `internal/services/calendar/`

Converts `models.Lesson` values into an ICS calendar using `github.com/arran4/golang-ical`.

Each event contains:

- `SUMMARY`: lesson name.
- `DTSTART` and `DTEND`: parsed lesson time.
- `DESCRIPTION`: online link or information from the announcement.
- `URL`: present when a call/course URL was found.

Calendar generation is sequential. Do not add concurrent writes to `ics.Calendar`; it is not thread-safe. A previous concurrency attempt was reverted (`b1c62f5` added it, `fd970d6` removed it).

### `internal/usecase/`

- `index`: returns available timetable file metadata as JSON.
- `get-calendar`: downloads schedules, extracts group lessons, loads online links, enriches lessons, and serializes the calendar.

### `internal/handlers/`

Thin `net/http` handlers for the index and calendar endpoints.

### `internal/models/`

Shared request, response, timetable file, lesson, and online announcement structures.

## Current Hardcoded Assumptions

These values are currently hardcoded and should be treated carefully when changing date behavior:

- Timezone: `Europe/Ulyanovsk`.
- Lesson start time: `18:00`.
- Lesson duration: 3 hours (`18:00`–`21:00`).
- Excel date format: `02.Jan`.
- Academic year: `2026`, appended while parsing each Excel date.
- Online page URL: `https://coe.ulstu.ru/index.php?action=show_page&id=103`.
- HTTP port: environment variable `PORT`, default `8589`.

The fixed year is a known limitation. Do not silently replace it with `time.Now().Year()` without considering schedules spanning calendar years.

## Tests

Tests currently cover:

- Safe handling of short/truncated Excel rows.
- Lesson date parsing.
- Online announcement parsing, including multiple dates/subjects and missing URLs.
- Extracting hidden anchor URLs.
- Subject normalization and lesson-link matching.
- Newest-announcement precedence.
- ICS `URL` and `DESCRIPTION` generation.

Run focused package tests first when changing a specific layer, then run:

```bash
go test ./...
go build -o /tmp/timetable-to-ics ./cmd
git diff --check
```

Live page parsing can change when ULSTU editors alter markup. For parser changes, verify against the real page in addition to fixture-based unit tests.

## Key Dependencies

- `github.com/xuri/excelize/v2` — Excel parsing.
- `github.com/arran4/golang-ical` — ICS generation.
- `github.com/google/uuid` — event UIDs.
- `golang.org/x/net/html` — HTML parsing.
- `golang.org/x/text/encoding/charmap` — Windows-1251 decoding.
- `time/tzdata` — embedded timezone database for environments without system timezone data.

## Engineering Notes

- Keep ULSTU-specific parsing in `internal/clients/ulstu`.
- Keep subject/date matching and lesson enrichment in `internal/services/lesson`.
- Keep ICS-specific behavior in `internal/services/calendar`.
- Preserve graceful handling of malformed or shortened Excel rows.
- Prefer parser fixtures for unusual HTML layouts and retain live-page validation for integration confidence.
- Do not modify unrelated local changes in the working tree.
