FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /timetable-to-ics ./cmd/

FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /timetable-to-ics /app/timetable-to-ics

ENV PORT=8589

EXPOSE ${PORT}

ENTRYPOINT ["/app/timetable-to-ics"]