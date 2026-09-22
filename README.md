# EOS Open Storage - public site

A modern replacement for [eos.web.cern.ch](https://eos.web.cern.ch): one Go process, SQLite, and a locked controller.

| | |
| --- | --- |
| Public site | `http://localhost:8080/` |
| Controller | `http://localhost:8080/controller` |

Without a certificate it listens on `:8080`. With `TLS_CERT` it defaults to `:443`.

## Quick start

You need [Go](https://go.dev/dl/) 1.26+.

```bash
export CONTROLLER_SECRET=dev-secret
make run
```

Open [http://localhost:8080/](http://localhost:8080/) and [http://localhost:8080/controller](http://localhost:8080/controller).

```bash
go test ./...
```

## Features

### Site-wide

- Single-page app with fade navigation between `/`, About, Tech, Roadmap, Docs, Workshops, Search talks, Resources, Service, News, and Community
- Sticky black header with Home icon, section links, GitLab, GitHub, and the CERN mark
- Footer with CERN Storage & Data Management Group address, support email, and a newsletter subscribe form
- Contact form (stored for the controller inbox)
- Links to [eos-docs.web.cern.ch](https://eos-docs.web.cern.ch/diopside/) open in an **on-site overlay**: the Sphinx article is fetched, restyled to the EOS palette, and previous/next chapters stay in the viewer. Cmd/Ctrl-click still opens the original page. Escape or the backdrop closes it
- Floating **Ask EOS** chat (see below)
- Reduced-motion preferences disable decorative animation

### Home

- Hero with kicker, title, lede, and four actions (workshop, latest version, install, search talks)
- Interactive CERN raw-disk capacity chart (2010–2025, one point per year). After 2025 the path is an open band that only pins the 2.5 EB target in 2030
- Copy-to-clipboard `git clone` command with a typewriter prompt
- Rotating EOS hexagon logo with an inward wind swirl, above the orbit movie
- Scrolling ticker of the latest news item under the hero buttons
- Scale stats: storage volume, IO, disks, files, clients
- About blurb and feature cards (flexible, scale, CERN, security, CERNBox, tape)
- “Help from the docs” cards that open the overlay viewer

### About

- Longer About copy and the same feature cards, with a cycling highlight on hover

### Tech

- Design & architecture narrative (MGM, FST, QuarkDB, XRootD)
- Interactive architecture figure and a four-step I/O flow
- Core service cards, cluster organisation (views, virtual identities, policies, GEO), and microservices (balancers, converter, lifecycle, LRU, inspector, consistency, workflow)
- Links into the Diopside docs (overlay) plus GitHub, GitLab, community, JIRA, CI, and RPMs

### Roadmap

- Timeline of EOS history and planned work toward later lines (including EOS 6)

### Documentation

- Curated “start here” topics from the Diopside manual (intro, architecture, install, CLI, auth, releases)
- In-page search over the extracted manual
- Cards and search hits open the overlay viewer instead of leaving the site

### Workshops

- All EOS workshops on Indico from 2018–2026 (2nd–10th), with dates, venue, and Indico links
- Jump to the talks search

### Search talks

- Full-text search (SQLite FTS5) over workshop talks and the docs extract
- Filters: talks, docs, or both, and by year
- Results include Indico, slides, and recordings when available

### Resources

- Documentation, workshop talks, presentations, publications, release notes, search commits
- Search commits: live git log of the EOS `master` branch (headings, messages, authors)
- Publications list with DOI / CERN CDS links

### Service

- CERN-facing services with previews: Control Tower, EOS Orbit, CERNBox, SWAN, CTA, CERN IT status

### News

- Dated news items (workshops, releases, CERN stories). Latest item also drives the home ticker

### Community

- Community forum / site links, collaborations, and the people roster
- Support pointers (eos-support, JIRA)

### Ask EOS

- Floating chat on every page
- Answers from live web search by default (Bing, then DuckDuckGo HTML). Optional official Google CSE if `GOOGLE_CSE_*` is set
- Chrome / Google AI Mode only when `CHAT_SEARCH_CHROME=1` or `CHAT_SEARCH_HEADED=1` (often captcha-blocked)
- Optional Gemini 2.5 Flash rewrite when an API key is set
- Source links, enlarge, and clear

## Controller

Locked admin at `/controller`. Tabs: **Home**, **About**, **Tech**, **News**, **Team**, **Index**, **Inbox**.

Hero copy, stats, cards, and people are editable. The **Index** tab can refresh Indico and eos-docs live, or reload the bundled seed. Newsletter signups and contact messages land in **Inbox**.

## Environment

| Variable | Default | Meaning |
| --- | --- | --- |
| `CONTROLLER_SECRET` | _(empty)_ | Shared secret for `/controller` |
| `GEMINI_API_KEY` / `GOOGLE_API_KEY` | _(empty)_ | Optional Gemini rewrite of the Ask EOS answer |
| `CHAT_SEARCH_HEADED` | _(empty)_ | `1` opens a visible Chrome window for Google AI Mode |
| `CHAT_SEARCH_CHROME` | _(empty)_ | `1` also tries headless Chrome Google (often captcha-blocked) |
| `CHAT_CHROME_PROFILE` | `data/chrome-profile` | Persistent Chrome profile (cookies / consent) |
| `CHAT_CHROME_BIN` | _(auto)_ | Chrome/Chromium binary on remote hosts |
| `GOOGLE_CSE_ID` / `GOOGLE_CSE_KEY` | _(empty)_ | Official Google Programmable Search (works without Chrome) |
| `ADDR` | `:8080` / `:443` with TLS | Listen address |
| `DATA_DIR` | `data` | SQLite + uploads (`data/eos.db`) |
| `TLS_CERT` | _(empty)_ | Certificate PEM; enables HTTPS |
| `TLS_KEY` | _(empty)_ | Private key PEM |

## Layout

```
cmd/server/          HTTP(S) server
internal/api/        Public + controller routes (including /api/docs/view)
internal/store/      SQLite + FTS5 search
internal/ingest/     Indico + eos-docs fetchers
internal/docsview/   Fetch and restyle eos-docs articles for the overlay
internal/googleai/   Live HTML search, optional CSE, optional Chrome
internal/chat/       Ask EOS
web/public/          Public UI (embedded)
web/controller/      Controller UI (embedded)
```

Restart the server after you change `web/` for Go-embedded assets. Disk fallback serves `web/public` directly, so CSS/JS tweaks show without a rebuild.
