# EOS Open Storage — public site

A modern replacement for [eos.web.cern.ch](https://eos.web.cern.ch): one Go process, SQLite, and a locked controller — the same shape as the New Spirit homepage.

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

## What the public site does

- Home, about, architecture, CERN services, news, community
- Documentation topics extracted from [eos-docs.web.cern.ch](https://eos-docs.web.cern.ch/diopside/)
- All EOS workshops on Indico from 2018–2026 (2nd–10th), with slides and recordings
- Full-text search over talks and the docs extract
- Newsletter and contact form (stored for the controller)
- Floating **Ask EOS** chat (Gemini 2.5 Flash + Chrome Google search)

## Controller

Tabs: **Home**, **About**, **Tech**, **News**, **Team**, **Index**, **Inbox**.

Hero copy, stats, cards, and people are editable. The **Index** tab can refresh Indico and eos-docs live, or reload the bundled seed.

## Environment

| Variable | Default | Meaning |
| --- | --- | --- |
| `CONTROLLER_SECRET` | _(empty)_ | Shared secret for `/controller` |
| `GEMINI_API_KEY` / `GOOGLE_API_KEY` | _(empty)_ | Optional Gemini rewrite of the Ask EOS answer |
| `CHAT_SEARCH_HEADED` | _(empty)_ | `1` opens a visible Chrome window for Google AI Mode |
| `CHAT_CHROME_PROFILE` | `data/chrome-profile` | Persistent Chrome profile (cookies / consent) |
| `ADDR` | `:8080` / `:443` with TLS | Listen address |
| `DATA_DIR` | `data` | SQLite + uploads (`data/eos.db`) |
| `TLS_CERT` | _(empty)_ | Certificate PEM; enables HTTPS |
| `TLS_KEY` | _(empty)_ | Private key PEM |

## Layout

```
cmd/server/          HTTP(S) server
internal/api/        Public + controller routes
internal/store/      SQLite + FTS5 search
internal/ingest/     Indico + eos-docs fetchers
internal/googleai/   Chrome → Google Search → AI Mode
web/public/          Public UI (embedded)
web/controller/      Controller UI (embedded)
```

Restart the server after you change `web/`.
