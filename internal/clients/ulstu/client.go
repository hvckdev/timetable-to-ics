package ulstu

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"
)

const (
	baseURL         = "https://coe.ulstu.ru"
	schedulePageURL = "/index.php?action=show_page&id=103"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type ScheduleFile struct {
	DisplayName string
	URL         string
	Month       time.Month
	Year        int
}

// ListLatestSchedules returns the most recently added schedule file for each month.
func (c *Client) ListLatestSchedules(ctx context.Context) ([]ScheduleFile, error) {
	files, err := c.listScheduleFiles(ctx)
	if err != nil {
		return nil, err
	}

	latest := make(map[string]ScheduleFile)
	for _, f := range files {
		key := fmt.Sprintf("%d-%d", f.Year, f.Month)
		latest[key] = f // last occurrence wins — files are in upload order
	}

	result := make([]ScheduleFile, 0, len(latest))
	for _, f := range latest {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Year != result[j].Year {
			return result[i].Year < result[j].Year
		}
		return result[i].Month < result[j].Month
	})

	return result, nil
}

// Download fetches the schedule file content as bytes.
func (c *Client) Download(ctx context.Context, sf ScheduleFile) ([]byte, error) {
	url := sf.URL
	if !strings.HasPrefix(url, "http") {
		url = baseURL + url
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading file: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return data, nil
}

func (c *Client) listScheduleFiles(ctx context.Context) ([]ScheduleFile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+schedulePageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching page: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	reader := charmap.Windows1251.NewDecoder().Reader(resp.Body)
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parsing html: %w", err)
	}

	var files []ScheduleFile
	inAttachments := false
	attachmentsDone := false

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "p" && hasClass(n, "page_title") && !attachmentsDone {
				text := strings.TrimSpace(getTextContent(n))
				if text == "Приложения:" {
					inAttachments = true
				}
				return
			}

			if inAttachments && n.Data == "a" && hasClass(n, "list_view") {
				href := getAttr(n, "href")
				displayName := strings.TrimSpace(getTextContent(n))

				if isScheduleXLSX(displayName, href) {
					months, year := parseMonthYear(displayName)
					for _, m := range months {
						files = append(files, ScheduleFile{
							DisplayName: displayName,
							URL:         baseURL + href,
							Month:       m,
							Year:        year,
						})
					}
				}
				return
			}
		}

		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			traverse(ch)
		}

		if inAttachments && n.Type == html.ElementNode && n.Data == "ul" {
			inAttachments = false
			attachmentsDone = true
		}
	}

	traverse(doc)
	return files, nil
}

func isScheduleXLSX(displayName, href string) bool {
	if href == "" || href == "/userfiles/image/" {
		return false
	}
	if !strings.HasSuffix(strings.ToLower(href), ".xlsx") {
		return false
	}
	return strings.Contains(strings.ToLower(displayName), "расписание")
}

var russianMonths = map[string]time.Month{
	"январь":   time.January,
	"февраль":  time.February,
	"март":     time.March,
	"апрель":   time.April,
	"май":      time.May,
	"июнь":     time.June,
	"июль":     time.July,
	"август":   time.August,
	"сентябрь": time.September,
	"октябрь":  time.October,
	"ноябрь":   time.November,
	"декабрь":  time.December,
}

func parseMonthYear(displayName string) ([]time.Month, int) {
	lower := strings.ToLower(displayName)

	var months []time.Month
	for ru, m := range russianMonths {
		if strings.Contains(lower, ru) {
			months = append(months, m)
		}
	}

	year := 0
	parts := strings.FieldsFunc(displayName, func(r rune) bool {
		return r == '_' || r == ' ' || r == '('
	})
	for _, p := range parts {
		if len(p) == 2 {
			y, err := strconv.Atoi(p)
			if err == nil && y >= 0 && y <= 99 {
				year = 2000 + y
			}
		}
	}

	return months, year
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			for _, c := range strings.Fields(attr.Val) {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getTextContent(c))
	}
	return sb.String()
}
