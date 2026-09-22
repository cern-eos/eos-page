package store

import (
	"strconv"
	"strings"
)

type Page struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	UpdatedAt string `json:"updatedAt"`
}

type Card struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Sort    int    `json:"sort"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Href    string `json:"href"`
	Meta    string `json:"meta"`
	Visible bool   `json:"visible"`
}

type News struct {
	ID        string `json:"id"`
	Sort      int    `json:"sort"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	DateLabel string `json:"dateLabel"`
	Href      string `json:"href"`
	Visible   bool   `json:"visible"`
}

type Person struct {
	ID      string `json:"id"`
	Sort    int    `json:"sort"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	Email   string `json:"email"`
	Visible bool   `json:"visible"`
}

type Workshop struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Year        int    `json:"year"`
	Edition     int    `json:"edition"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Location    string `json:"location"`
	Room        string `json:"room"`
	Kind        string `json:"kind"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Visible     bool   `json:"visible"`
}

type Talk struct {
	ID        string `json:"id"`
	EventID   string `json:"eventId"`
	Year      int    `json:"year"`
	Title     string `json:"title"`
	Abstract  string `json:"abstract"`
	Speakers  string `json:"speakers"`
	Session   string `json:"session"`
	URL       string `json:"url"`
	Slides    string `json:"slides"`
	Recording string `json:"recording"`
	Start     string `json:"start"`
}

type Doc struct {
	ID      string `json:"id"`
	PageID  string `json:"pageId"`
	Title   string `json:"title"`
	Section string `json:"section"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
	Body    string `json:"body"`
	Sort    int    `json:"sort"`
}

type Subscriber struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
	Status    string `json:"status"`
}

type InboxMsg struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

func (s *Store) Page(id string) (Page, error) {
	var p Page
	err := s.db.QueryRow(`SELECT id, title, body, updated_at FROM pages WHERE id=?`, id).
		Scan(&p.ID, &p.Title, &p.Body, &p.UpdatedAt)
	if err != nil {
		return Page{}, ErrNotFound
	}
	return p, nil
}

func (s *Store) ListPages() ([]Page, error) {
	rows, err := s.db.Query(`SELECT id, title, body, updated_at FROM pages ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Page
	for rows.Next() {
		var p Page
		if err := rows.Scan(&p.ID, &p.Title, &p.Body, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpsertPage(p Page) error {
	if p.ID == "" {
		return ErrInvalid
	}
	_, err := s.db.Exec(`
INSERT INTO pages(id, title, body, updated_at) VALUES(?,?,?,?)
ON CONFLICT(id) DO UPDATE SET title=excluded.title, body=excluded.body, updated_at=excluded.updated_at`,
		p.ID, p.Title, p.Body, nowISO())
	return err
}

func (s *Store) ListCards(kind string, publicOnly bool) ([]Card, error) {
	q := `SELECT id, kind, sort, title, body, href, meta, visible FROM cards`
	args := []any{}
	where := []string{}
	if kind != "" {
		where = append(where, `kind=?`)
		args = append(args, kind)
	}
	if publicOnly {
		where = append(where, `visible=1`)
	}
	if len(where) > 0 {
		q += ` WHERE ` + strings.Join(where, " AND ")
	}
	q += ` ORDER BY kind, sort, title`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Card
	for rows.Next() {
		var c Card
		var vis int
		if err := rows.Scan(&c.ID, &c.Kind, &c.Sort, &c.Title, &c.Body, &c.Href, &c.Meta, &vis); err != nil {
			return nil, err
		}
		c.Visible = vis == 1
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) UpsertCard(c Card) error {
	if c.ID == "" {
		c.ID = NewID()
	}
	if c.Kind == "" {
		return ErrInvalid
	}
	_, err := s.db.Exec(`
INSERT INTO cards(id, kind, sort, title, body, href, meta, visible)
VALUES(?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  kind=excluded.kind, sort=excluded.sort, title=excluded.title, body=excluded.body,
  href=excluded.href, meta=excluded.meta, visible=excluded.visible`,
		c.ID, c.Kind, c.Sort, c.Title, c.Body, c.Href, c.Meta, boolInt(c.Visible))
	return err
}

func (s *Store) DeleteCard(id string) error {
	res, err := s.db.Exec(`DELETE FROM cards WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListNews(publicOnly bool) ([]News, error) {
	q := `SELECT id, sort, title, body, date_label, href, visible FROM news`
	if publicOnly {
		q += ` WHERE visible=1`
	}
	q += ` ORDER BY sort, title`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []News
	for rows.Next() {
		var n News
		var vis int
		if err := rows.Scan(&n.ID, &n.Sort, &n.Title, &n.Body, &n.DateLabel, &n.Href, &vis); err != nil {
			return nil, err
		}
		n.Visible = vis == 1
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) UpsertNews(n News) error {
	if n.ID == "" {
		n.ID = NewID()
	}
	_, err := s.db.Exec(`
INSERT INTO news(id, sort, title, body, date_label, href, visible)
VALUES(?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  sort=excluded.sort, title=excluded.title, body=excluded.body,
  date_label=excluded.date_label, href=excluded.href, visible=excluded.visible`,
		n.ID, n.Sort, n.Title, n.Body, n.DateLabel, n.Href, boolInt(n.Visible))
	return err
}

func (s *Store) DeleteNews(id string) error {
	res, err := s.db.Exec(`DELETE FROM news WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListPeople(publicOnly bool) ([]Person, error) {
	q := `SELECT id, sort, name, role, email, visible FROM people`
	if publicOnly {
		q += ` WHERE visible=1`
	}
	q += ` ORDER BY sort, name`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Person
	for rows.Next() {
		var p Person
		var vis int
		if err := rows.Scan(&p.ID, &p.Sort, &p.Name, &p.Role, &p.Email, &vis); err != nil {
			return nil, err
		}
		p.Visible = vis == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpsertPerson(p Person) error {
	if p.ID == "" {
		p.ID = NewID()
	}
	_, err := s.db.Exec(`
INSERT INTO people(id, sort, name, role, email, visible)
VALUES(?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  sort=excluded.sort, name=excluded.name, role=excluded.role, email=excluded.email, visible=excluded.visible`,
		p.ID, p.Sort, p.Name, p.Role, p.Email, boolInt(p.Visible))
	return err
}

func (s *Store) DeletePerson(id string) error {
	res, err := s.db.Exec(`DELETE FROM people WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListWorkshops(publicOnly bool) ([]Workshop, error) {
	q := `SELECT id, title, year, edition, start_date, end_date, location, room, kind, url, description, sort, visible FROM workshops`
	if publicOnly {
		q += ` WHERE visible=1`
	}
	q += ` ORDER BY year DESC, edition DESC`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Workshop
	for rows.Next() {
		var w Workshop
		var vis int
		if err := rows.Scan(&w.ID, &w.Title, &w.Year, &w.Edition, &w.Start, &w.End, &w.Location, &w.Room, &w.Kind, &w.URL, &w.Description, &w.Sort, &vis); err != nil {
			return nil, err
		}
		w.Visible = vis == 1
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *Store) UpsertWorkshop(w Workshop) error {
	if w.ID == "" {
		return ErrInvalid
	}
	_, err := s.db.Exec(`
INSERT INTO workshops(id, title, year, edition, start_date, end_date, location, room, kind, url, description, sort, visible)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  title=excluded.title, year=excluded.year, edition=excluded.edition, start_date=excluded.start_date,
  end_date=excluded.end_date, location=excluded.location, room=excluded.room, kind=excluded.kind,
  url=excluded.url, description=excluded.description, sort=excluded.sort, visible=excluded.visible`,
		w.ID, w.Title, w.Year, w.Edition, w.Start, w.End, w.Location, w.Room, w.Kind, w.URL, w.Description, w.Sort, boolInt(w.Visible))
	return err
}

func (s *Store) ReplaceTalks(talks []Talk) error {
	uniqueIDs(talks, func(t Talk) string { return t.ID }, func(t *Talk, id string) { t.ID = id })
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM talks_fts`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM talks`); err != nil {
		return err
	}
	ins, err := tx.Prepare(`INSERT INTO talks(id, event_id, year, title, abstract, speakers, session, url, slides, recording, start_date)
VALUES(?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer ins.Close()
	fts, err := tx.Prepare(`INSERT INTO talks_fts(id, title, abstract, speakers, session, year) VALUES(?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer fts.Close()
	for _, t := range talks {
		if t.ID == "" {
			continue
		}
		if _, err := ins.Exec(t.ID, t.EventID, t.Year, t.Title, t.Abstract, t.Speakers, t.Session, t.URL, t.Slides, t.Recording, t.Start); err != nil {
			return err
		}
		if _, err := fts.Exec(t.ID, t.Title, t.Abstract, t.Speakers, t.Session, t.Year); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func uniqueIDs[T any](items []T, get func(T) string, set func(*T, string)) {
	seen := map[string]int{}
	for i := range items {
		id := get(items[i])
		if id == "" {
			id = NewID()
		}
		base := id
		for n := 0; ; n++ {
			cand := base
			if n > 0 {
				cand = base + "-" + strconv.Itoa(n)
			}
			if seen[cand] == 0 {
				set(&items[i], cand)
				seen[cand] = 1
				break
			}
		}
	}
}

func (s *Store) ReplaceDocs(docs []Doc) error {
	uniqueIDs(docs, func(d Doc) string { return d.ID }, func(d *Doc, id string) { d.ID = id })
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM docs_fts`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM docs`); err != nil {
		return err
	}
	ins, err := tx.Prepare(`INSERT INTO docs(id, page_id, title, section, url, summary, body, sort) VALUES(?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer ins.Close()
	fts, err := tx.Prepare(`INSERT INTO docs_fts(id, title, section, summary, body) VALUES(?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer fts.Close()
	for i, d := range docs {
		if d.ID == "" {
			continue
		}
		if d.Sort == 0 {
			d.Sort = i
		}
		if _, err := ins.Exec(d.ID, d.PageID, d.Title, d.Section, d.URL, d.Summary, d.Body, d.Sort); err != nil {
			return err
		}
		if _, err := fts.Exec(d.ID, d.Title, d.Section, d.Summary, d.Body); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) TalkCounts() (talks, slides, recordings int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*), SUM(CASE WHEN slides!='' THEN 1 ELSE 0 END), SUM(CASE WHEN recording!='' THEN 1 ELSE 0 END) FROM talks`).
		Scan(&talks, &slides, &recordings)
	return
}

func (s *Store) DocCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM docs`).Scan(&n)
	return n, err
}

func (s *Store) FeaturedDocs(limit int) ([]Doc, error) {
	if limit <= 0 {
		limit = 12
	}
	rows, err := s.db.Query(`
SELECT d.id, d.page_id, d.title, d.section, d.url, d.summary, d.body, d.sort
FROM docs d
JOIN (SELECT page_id, MIN(rowid) AS rid FROM docs GROUP BY page_id) x ON d.rowid = x.rid
ORDER BY
  CASE d.page_id
    WHEN 'intro' THEN 1
    WHEN 'architecture' THEN 2
    WHEN 'getting-started' THEN 3
    WHEN 'using' THEN 4
    WHEN 'configuration' THEN 5
    WHEN 'protocols' THEN 6
    WHEN 'microservices' THEN 7
    WHEN 'diopside-release' THEN 8
    WHEN 'faq' THEN 9
    ELSE 50
  END, d.title
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Doc
	for rows.Next() {
		var d Doc
		if err := rows.Scan(&d.ID, &d.PageID, &d.Title, &d.Section, &d.URL, &d.Summary, &d.Body, &d.Sort); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) Subscribe(email string) (Subscriber, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") {
		return Subscriber{}, ErrInvalid
	}
	var existing Subscriber
	err := s.db.QueryRow(`SELECT id, email, created_at, status FROM subscribers WHERE email=?`, email).
		Scan(&existing.ID, &existing.Email, &existing.CreatedAt, &existing.Status)
	if err == nil {
		return existing, nil
	}
	sub := Subscriber{ID: NewID(), Email: email, CreatedAt: nowISO(), Status: "pending"}
	_, err = s.db.Exec(`INSERT INTO subscribers(id, email, created_at, status) VALUES(?,?,?,?)`,
		sub.ID, sub.Email, sub.CreatedAt, sub.Status)
	return sub, err
}

func (s *Store) ListSubscribers() ([]Subscriber, error) {
	rows, err := s.db.Query(`SELECT id, email, created_at, status FROM subscribers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subscriber
	for rows.Next() {
		var sub Subscriber
		if err := rows.Scan(&sub.ID, &sub.Email, &sub.CreatedAt, &sub.Status); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (s *Store) SetSubscriberStatus(id, status string) (Subscriber, error) {
	_, err := s.db.Exec(`UPDATE subscribers SET status=? WHERE id=?`, status, id)
	if err != nil {
		return Subscriber{}, err
	}
	var sub Subscriber
	err = s.db.QueryRow(`SELECT id, email, created_at, status FROM subscribers WHERE id=?`, id).
		Scan(&sub.ID, &sub.Email, &sub.CreatedAt, &sub.Status)
	if err != nil {
		return Subscriber{}, ErrNotFound
	}
	return sub, nil
}

func (s *Store) AddInbox(kind, name, email, message string) (InboxMsg, error) {
	if strings.TrimSpace(name) == "" || !strings.Contains(email, "@") {
		return InboxMsg{}, ErrInvalid
	}
	msg := InboxMsg{ID: NewID(), Kind: kind, Name: name, Email: email, Message: message, CreatedAt: nowISO()}
	_, err := s.db.Exec(`INSERT INTO inbox(id, kind, name, email, message, created_at) VALUES(?,?,?,?,?,?)`,
		msg.ID, msg.Kind, msg.Name, msg.Email, msg.Message, msg.CreatedAt)
	return msg, err
}

func (s *Store) ListInbox() ([]InboxMsg, error) {
	rows, err := s.db.Query(`SELECT id, kind, name, email, message, created_at FROM inbox ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InboxMsg
	for rows.Next() {
		var m InboxMsg
		if err := rows.Scan(&m.ID, &m.Kind, &m.Name, &m.Email, &m.Message, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
