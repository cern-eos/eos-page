package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/apeters/eospage/internal/store"
)

const defaultUA = "eos-page/1.0 (+https://eos.web.cern.ch)"

var DefaultEventIDs = []string{
	"1622471", "1483930", "1353101", "1227241", "1103358",
	"985953", "862873", "775181", "656157",
}

type indicoExport struct {
	Results []indicoEvent `json:"results"`
}

type indicoEvent struct {
	ID           any               `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	URL          string            `json:"url"`
	Location     string            `json:"location"`
	Room         string            `json:"room"`
	RoomFullname string            `json:"roomFullname"`
	StartDate    indicoDate        `json:"startDate"`
	EndDate      indicoDate        `json:"endDate"`
	Contributions []indicoContrib `json:"contributions"`
}

type indicoDate struct {
	Date string `json:"date"`
}

type indicoContrib struct {
	ID             any            `json:"id"`
	DBID           any            `json:"db_id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	URL            string         `json:"url"`
	Session        string         `json:"session"`
	StartDate      indicoDate     `json:"startDate"`
	Speakers       []indicoPerson `json:"speakers"`
	PrimaryAuthors []indicoPerson `json:"primaryauthors"`
	Folders        []indicoFolder `json:"folders"`
}

type indicoPerson struct {
	FullName    string `json:"fullName"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Affiliation string `json:"affiliation"`
}

type indicoFolder struct {
	Attachments []indicoAttach `json:"attachments"`
}

type indicoAttach struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	DownloadURL string `json:"download_url"`
	LinkURL     string `json:"link_url"`
}

func ParseEventIDs(raw string) []string {
	seen := map[string]bool{}
	var out []string
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';' || r == '\n'
	}) {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		if _, err := strconv.Atoi(part); err != nil {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	if len(out) == 0 {
		return append([]string{}, DefaultEventIDs...)
	}
	return out
}

func FetchWorkshops(ids []string) ([]store.Workshop, []store.Talk, error) {
	client := &http.Client{Timeout: 45 * time.Second}
	var workshops []store.Workshop
	var talks []store.Talk
	for _, id := range ids {
		w, ts, err := fetchEvent(client, id)
		if err != nil {
			return nil, nil, fmt.Errorf("indico %s: %w", id, err)
		}
		workshops = append(workshops, w)
		talks = append(talks, ts...)
	}
	return workshops, talks, nil
}

func fetchEvent(client *http.Client, id string) (store.Workshop, []store.Talk, error) {
	url := fmt.Sprintf("https://indico.cern.ch/export/event/%s.json?detail=contributions", id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return store.Workshop{}, nil, err
	}
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return store.Workshop{}, nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return store.Workshop{}, nil, err
	}
	if res.StatusCode != http.StatusOK {
		return store.Workshop{}, nil, fmt.Errorf("HTTP %d", res.StatusCode)
	}
	var payload indicoExport
	if err := json.Unmarshal(body, &payload); err != nil {
		return store.Workshop{}, nil, err
	}
	if len(payload.Results) == 0 {
		return store.Workshop{}, nil, fmt.Errorf("empty result")
	}
	ev := payload.Results[0]
	year := yearOf(ev.StartDate.Date)
	w := store.Workshop{
		ID:          id,
		Title:       ev.Title,
		Year:        year,
		Edition:     editionFor(year),
		Start:       ev.StartDate.Date,
		End:         ev.EndDate.Date,
		Location:    ev.Location,
		Room:        firstNonEmpty(ev.RoomFullname, ev.Room),
		Kind:        kindFor(ev.Location, ev.Title),
		URL:         firstNonEmpty(ev.URL, "https://indico.cern.ch/event/"+id+"/"),
		Description: stripHTML(ev.Description),
		Sort:        3000 - year,
		Visible:     true,
	}
	var talks []store.Talk
	for _, c := range ev.Contributions {
		title := strings.TrimSpace(c.Title)
		if title == "" {
			continue
		}
		slides, rec := materials(c)
		talks = append(talks, store.Talk{
			ID:        id + "-" + anyString(firstAny(c.DBID, c.ID)),
			EventID:   id,
			Year:      year,
			Title:     title,
			Abstract:  clip(stripHTML(c.Description), 2500),
			Speakers:  peopleLine(c.Speakers, c.PrimaryAuthors),
			Session:   c.Session,
			URL:       c.URL,
			Slides:    slides,
			Recording: rec,
			Start:     c.StartDate.Date,
		})
	}
	return w, talks, nil
}

func editionFor(year int) int {
	if year >= 2018 && year <= 2030 {
		return year - 2016
	}
	return 0
}

func kindFor(location, title string) string {
	s := strings.ToLower(location + " " + title)
	if strings.Contains(s, "virtual") || strings.Contains(s, "online") {
		return "virtual"
	}
	return "in-person"
}

func yearOf(date string) int {
	if len(date) >= 4 {
		if y, err := strconv.Atoi(date[:4]); err == nil {
			return y
		}
	}
	return 0
}

func materials(c indicoContrib) (slides, recording string) {
	for _, folder := range c.Folders {
		for _, att := range folder.Attachments {
			url := firstNonEmpty(att.DownloadURL, att.LinkURL)
			title := strings.ToLower(att.Title + " " + att.Filename)
			if att.Type == "link" && recording == "" && (strings.Contains(title, "record") || strings.Contains(url, "videos.cern.ch") || strings.Contains(title, "video")) {
				recording = firstNonEmpty(att.LinkURL, url)
				continue
			}
			fn := strings.ToLower(att.Filename)
			if slides == "" && (strings.HasSuffix(fn, ".pdf") || strings.HasSuffix(fn, ".pptx") || strings.HasSuffix(fn, ".ppt") || strings.Contains(att.ContentType, "pdf")) {
				slides = url
			}
		}
	}
	return
}

func peopleLine(a, b []indicoPerson) string {
	seen := map[string]bool{}
	var names []string
	for _, p := range append(append([]indicoPerson{}, a...), b...) {
		n := strings.TrimSpace(p.FullName)
		if n == "" {
			n = strings.TrimSpace(p.FirstName + " " + p.LastName)
		}
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		if p.Affiliation != "" {
			n += " (" + p.Affiliation + ")"
		}
		names = append(names, n)
	}
	return strings.Join(names, ", ")
}

var tagRe = regexp.MustCompile(`(?is)<[^>]+>`)

func stripHTML(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = tagRe.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	return strings.Join(strings.Fields(s), " ")
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func firstAny(v ...any) any {
	for _, x := range v {
		if x != nil && anyString(x) != "" && anyString(x) != "<nil>" {
			return x
		}
	}
	return ""
}

func anyString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprint(v)
	}
}
