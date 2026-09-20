package ulstu

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestParseLessonLinks(t *testing.T) {
	page := `
15 сентября в 18.00
Предмет: Экономика и управление проектами в IT-отрасли
Преподаватель: Рыбкина Мария Васильевна
НОВАЯ ссылка на предмет: https://telemost.yandex.ru/j/41652278012541
________________________________________________________________________________
8 сентября в 18.00
Предмет: Системы ИИ
Преподаватель: Чекина Александра Валерьевна
Ссылка:
________________________________________________________________________________
11 сентября в 18.00 / 15 сентября в 18.00
Предметы: Методы и алгоритмы конвертации данных / Технологии создания человеко-машинного интерфейса
Преподаватель: Шамшев Анатолий Борисович
Ссылка: https://example.com/course
`

	links := parseLessonLinks(page)
	if len(links) != 4 {
		t.Fatalf("parseLessonLinks() returned %d links, want 4: %#v", len(links), links)
	}

	if links[0].Day != 15 || links[0].Month != time.September || links[0].URL != "https://telemost.yandex.ru/j/41652278012541" {
		t.Errorf("first link = %#v", links[0])
	}
	if links[1].Subject != "Системы ИИ" || links[1].URL != "" {
		t.Errorf("empty announcement = %#v", links[1])
	}
	if links[2].Day != 11 || links[2].Subject != "Методы и алгоритмы конвертации данных" {
		t.Errorf("first paired announcement = %#v", links[2])
	}
	if links[3].Day != 15 || links[3].Subject != "Технологии создания человеко-машинного интерфейса" {
		t.Errorf("second paired announcement = %#v", links[3])
	}
}

func TestGetTextContentWithLinksUsesHref(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`<p>Ссылка: <a href="https://example.com/call">подключиться</a></p>`))
	if err != nil {
		t.Fatal(err)
	}

	text := getTextContentWithLinks(doc)
	if !strings.Contains(text, "https://example.com/call") {
		t.Errorf("getTextContentWithLinks() = %q", text)
	}
}
