BINARY_NAME=timetable-to-ics
VERSION=1.0.0
BUILD=`date +%FT%T%z`

# Путь к главному файлу
MAIN_FILE=cmd/main.go

# Команды по умолчанию
all: windows linux darwin

# Сборка для Windows
windows:
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	GOOS=windows GOARCH=386 go build -o bin/$(BINARY_NAME)-windows-386.exe $(MAIN_FILE)

# Сборка для Linux
linux:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=386 go build -o bin/$(BINARY_NAME)-linux-386 $(MAIN_FILE)
	GOOS=linux GOARCH=arm go build -o bin/$(BINARY_NAME)-linux-arm $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)

# Сборка для macOS (Intel и ARM)
darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)

# Очистка сборочных файлов
clean:
	rm -rf bin/$(BINARY_NAME)-*

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make windows   - Сборка для Windows"
	@echo "  make linux     - Сборка для Linux"
	@echo "  make darwin    - Сборка для macOS (Intel и ARM)"
	@echo "  make clean     - Очистить сборочные файлы"
	@echo "  make all       - Выполнить все вышеуказанные команды"
