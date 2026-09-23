package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalid      = errors.New("invalid")
)

type Store struct {
	db       *sql.DB
	mediaDir string
}

func Open(dbPath, mediaDir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, mediaDir: mediaDir}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.seedIfEmpty(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureIndex(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureHelpCards(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureNewsWorkshop2027(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureTechContent(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureRoadmapContent(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensurePublicationsCard(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureCommunityRoster(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureOrbitService(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureSearchCommits(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.EnsureExternalPresentations(); err != nil {
		db.Close()
		return nil, err
	}
	if s.Setting("hero_video") == "" {
		if err := s.SetSetting("hero_video", "ttSjYYBOlsM"); err != nil {
			db.Close()
			return nil, err
		}
	}
	if s.Setting("stat_volume") == "930 PB" {
		if err := s.SetSetting("stat_volume", "1.1 EB"); err != nil {
			db.Close()
			return nil, err
		}
	}
	if s.Setting("stat_disks") == "70k" {
		if err := s.SetSetting("stat_disks", "100k"); err != nil {
			db.Close()
			return nil, err
		}
	}
	if s.Setting("stat_io") == "" {
		if err := s.SetSetting("stat_io", "1–2 TB/s"); err != nil {
			db.Close()
			return nil, err
		}
		if err := s.SetSetting("stat_io_label", "IO"); err != nil {
			db.Close()
			return nil, err
		}
	}
	const tower = "https://monit-grafana.cern.ch/d/baff3c33-decb-4b91-a6bf-c0ba84bdcbe4/eos-user-monitoring?orgId=22&from=now-24h&to=now&timezone=browser&var-cluster=$__all&var-HTTP=$__all&var-GRIDFPT=$__all&var-XROOTD=$__all&var-FUSE=$__all"
	if old := s.Setting("control_tower"); old == "" || strings.Contains(old, "filer-carbon") || strings.Contains(old, "eos-control-tower") {
		if err := s.SetSetting("control_tower", tower); err != nil {
			db.Close()
			return nil, err
		}
	}
	if _, err := s.db.Exec(`UPDATE cards SET href=? WHERE id='s-tower' AND (href LIKE '%filer-carbon%' OR href LIKE '%eos-control-tower%')`, tower); err != nil {
		db.Close()
		return nil, err
	}
	if addr := s.Setting("address"); strings.Contains(addr, "IT-ST") {
		next := strings.Replace(addr, "CERN IT-ST", "CERN Storage & Data Management Group", 1)
		if err := s.SetSetting("address", next); err != nil {
			db.Close()
			return nil, err
		}
	}
	if s.Setting("contributors") != "" {
		if err := s.SetSetting("contributors", ""); err != nil {
			db.Close()
			return nil, err
		}
	}
	if err := s.stripEmDashes(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) stripEmDashes() error {
	stmts := []string{
		`UPDATE settings SET value = REPLACE(REPLACE(value, ' — ', ' - '), '—', '-') WHERE value LIKE '%—%'`,
		`UPDATE pages SET title = REPLACE(REPLACE(title, ' — ', ' - '), '—', '-'), body = REPLACE(REPLACE(body, ' — ', ' - '), '—', '-') WHERE title LIKE '%—%' OR body LIKE '%—%'`,
		`UPDATE cards SET title = REPLACE(REPLACE(title, ' — ', ' - '), '—', '-'), body = REPLACE(REPLACE(body, ' — ', ' - '), '—', '-') WHERE title LIKE '%—%' OR body LIKE '%—%'`,
		`UPDATE news SET title = REPLACE(REPLACE(title, ' — ', ' - '), '—', '-'), body = REPLACE(REPLACE(body, ' — ', ' - '), '—', '-') WHERE title LIKE '%—%' OR body LIKE '%—%'`,
		`UPDATE people SET role = REPLACE(REPLACE(role, ' — ', ' - '), '—', '-') WHERE role LIKE '%—%'`,
		`UPDATE workshops SET title = REPLACE(REPLACE(title, ' — ', ' - '), '—', '-'), description = REPLACE(REPLACE(description, ' — ', ' - '), '—', '-') WHERE title LIKE '%—%' OR description LIKE '%—%'`,
		`UPDATE talks SET title = REPLACE(REPLACE(title, ' — ', ' - '), '—', '-'), abstract = REPLACE(REPLACE(abstract, ' — ', ' - '), '—', '-'), speakers = REPLACE(REPLACE(speakers, ' — ', ' - '), '—', '-'), session = REPLACE(REPLACE(session, ' — ', ' - '), '—', '-') WHERE title LIKE '%—%' OR abstract LIKE '%—%' OR speakers LIKE '%—%' OR session LIKE '%—%'`,
		`UPDATE docs SET title = REPLACE(REPLACE(title, ' — ', ' - '), '—', '-'), section = REPLACE(REPLACE(section, ' — ', ' - '), '—', '-'), summary = REPLACE(REPLACE(summary, ' — ', ' - '), '—', '-'), body = REPLACE(REPLACE(body, ' — ', ' - '), '—', '-') WHERE title LIKE '%—%' OR section LIKE '%—%' OR summary LIKE '%—%' OR body LIKE '%—%'`,
		`UPDATE commits SET title = REPLACE(REPLACE(title, ' — ', ' - '), '—', '-'), body = REPLACE(REPLACE(body, ' — ', ' - '), '—', '-'), author = REPLACE(REPLACE(author, ' — ', ' - '), '—', '-') WHERE title LIKE '%—%' OR body LIKE '%—%' OR author LIKE '%—%'`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	if _, err := s.db.Exec(`DELETE FROM talks_fts`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`INSERT INTO talks_fts(id, title, abstract, speakers, session, year) SELECT id, title, abstract, speakers, session, year FROM talks`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM docs_fts`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`INSERT INTO docs_fts(id, title, section, summary, body) SELECT id, title, section, summary, body FROM docs`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM commits_fts`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`INSERT INTO commits_fts(sha, title, body, author) SELECT sha, title, body, author FROM commits`); err != nil {
		return err
	}
	return nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) MediaDir() string { return s.mediaDir }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS pages (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS cards (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  sort INTEGER NOT NULL DEFAULT 0,
  title TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  href TEXT NOT NULL DEFAULT '',
  meta TEXT NOT NULL DEFAULT '',
  visible INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS news (
  id TEXT PRIMARY KEY,
  sort INTEGER NOT NULL DEFAULT 0,
  title TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  date_label TEXT NOT NULL DEFAULT '',
  href TEXT NOT NULL DEFAULT '',
  visible INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS people (
  id TEXT PRIMARY KEY,
  sort INTEGER NOT NULL DEFAULT 0,
  name TEXT NOT NULL DEFAULT '',
  role TEXT NOT NULL DEFAULT '',
  email TEXT NOT NULL DEFAULT '',
  visible INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS workshops (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL DEFAULT '',
  year INTEGER NOT NULL DEFAULT 0,
  edition INTEGER NOT NULL DEFAULT 0,
  start_date TEXT NOT NULL DEFAULT '',
  end_date TEXT NOT NULL DEFAULT '',
  location TEXT NOT NULL DEFAULT '',
  room TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  sort INTEGER NOT NULL DEFAULT 0,
  visible INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS talks (
  id TEXT PRIMARY KEY,
  event_id TEXT NOT NULL DEFAULT '',
  year INTEGER NOT NULL DEFAULT 0,
  title TEXT NOT NULL DEFAULT '',
  abstract TEXT NOT NULL DEFAULT '',
  speakers TEXT NOT NULL DEFAULT '',
  session TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT '',
  slides TEXT NOT NULL DEFAULT '',
  recording TEXT NOT NULL DEFAULT '',
  start_date TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS docs (
  id TEXT PRIMARY KEY,
  page_id TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  section TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT '',
  summary TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  sort INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS subscribers (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending'
);
CREATE TABLE IF NOT EXISTS inbox (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  email TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS chats (
  id TEXT PRIMARY KEY,
  created_at TEXT NOT NULL,
  question TEXT NOT NULL DEFAULT '',
  answer TEXT NOT NULL DEFAULT '',
  model TEXT NOT NULL DEFAULT '',
  ip TEXT NOT NULL DEFAULT ''
);
CREATE VIRTUAL TABLE IF NOT EXISTS talks_fts USING fts5(
  id UNINDEXED, title, abstract, speakers, session, year UNINDEXED,
  tokenize='porter'
);
CREATE VIRTUAL TABLE IF NOT EXISTS docs_fts USING fts5(
  id UNINDEXED, title, section, summary, body,
  tokenize='porter'
);
CREATE TABLE IF NOT EXISTS commits (
  sha TEXT PRIMARY KEY,
  short TEXT NOT NULL DEFAULT '',
  date TEXT NOT NULL DEFAULT '',
  year INTEGER NOT NULL DEFAULT 0,
  author TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT ''
);
CREATE VIRTUAL TABLE IF NOT EXISTS commits_fts USING fts5(
  sha UNINDEXED, title, body, author,
  tokenize='porter'
);
`)
	return err
}

func (s *Store) Setting(key string) string {
	var v string
	if err := s.db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&v); err != nil {
		return ""
	}
	return v
}

func (s *Store) Settings() (map[string]string, error) {
	rows, err := s.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

func (s *Store) SetSetting(key, value string) error {
	if key == "" {
		return ErrInvalid
	}
	_, err := s.db.Exec(`INSERT INTO settings(key,value) VALUES(?,?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

func (s *Store) CreateControllerSession() (string, error) {
	id := NewID()
	_, err := s.db.Exec(`INSERT INTO sessions(id, created_at) VALUES(?,?)`, id, nowISO())
	return id, err
}

func (s *Store) ValidControllerSession(id string) bool {
	if id == "" {
		return false
	}
	var created string
	if err := s.db.QueryRow(`SELECT created_at FROM sessions WHERE id=?`, id).Scan(&created); err != nil {
		return false
	}
	t, err := time.Parse(time.RFC3339, created)
	if err != nil {
		return false
	}
	return time.Since(t) < 12*time.Hour
}

func (s *Store) RevokeControllerSession(id string) {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE id=?`, id)
}

func NewID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func nowISO() string { return time.Now().UTC().Format(time.RFC3339) }

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
