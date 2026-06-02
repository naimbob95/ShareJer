# Share Jer

Upload a file, get a shareable link + QR code. Optional per-file password, automatic expiry, and one-click delete. A Go backend (standard library) with a Svelte frontend.

## Stack

**Backend**
- Go (1.25+), standard library `net/http` (Go 1.22 method-based routing)
- SQLite (`modernc.org/sqlite`, pure Go — no CGO)
- bcrypt for optional passwords
- `go-qrcode` for QR generation
- Local disk storage, behind a `Storage` interface (swappable for S3/R2 later)
- Background sweeper goroutine for expired-file cleanup + graceful shutdown

**Frontend** (`frontend/`)
- Svelte 5 + Vite + Tailwind CSS v4 (a separate static SPA)
- Talks to the API via a same-origin proxy (Vite in dev, nginx in prod) — no CORS

## Quick start

### Option A — Docker (runs both services)

```bash
docker compose up -d --build
```

Open **http://localhost:5173**. nginx serves the frontend and proxies `/api` to the Go server. Data (SQLite + uploads) persists in the `sharejer_data` volume. Stop with `docker compose down` (add `-v` to wipe the data volume).

### Option B — Local dev (hot reload)

```bash
# terminal 1 — Go API on :8080
go run ./cmd/server

# terminal 2 — Svelte dev server on :5173 (proxies /api to :8080)
cd frontend && npm install && npm run dev
```

Open **http://localhost:5173**.

## Configuration

The server reads environment variables, optionally seeded from a `.env` file in the working directory (see `.env.example`). All values have defaults.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | API listen port |
| `BASE_URL` | `http://localhost:8080` | Public origin where share links + QR codes resolve. Set to the **frontend** origin (e.g. `http://localhost:5173`), since links point at the `/f/{id}` page. |
| `MAX_UPLOAD_SIZE_MB` | `5` | Max upload size in megabytes |
| `FILE_EXPIRY` | `15m` | How long a file stays reachable after upload (Go duration). `0` disables expiry. |
| `CLEANUP_INTERVAL` | `5m` | How often the background sweeper purges expired files (Go duration) |
| `DB_PATH` | `sharejer.db` | SQLite file path |
| `UPLOADS_DIR` | `uploads` | Directory for stored files |

## API

All endpoints are served by the Go backend (reached via the proxy at `/api/...`).

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Liveness check, returns `ok`. |
| GET | `/api/config` | Upload limits for the UI: `{ maxUploadBytes, maxUploadMB, expirySeconds }`. |
| POST | `/api/upload` | Multipart upload, optional `password` field. Returns `{ id, shareUrl }` where `shareUrl` is `BASE_URL/f/{id}`. |
| GET | `/api/file/{id}` | File metadata (filename, size, expiry, whether a password is required). Never exposes the hash. |
| POST | `/api/file/{id}/download` | Verify password if set, stream the file. |
| DELETE | `/api/file/{id}` | Delete the file (and its bytes). Requires the password in the JSON body if the file is protected. |
| GET | `/api/qr/{id}` | PNG QR code encoding the share URL. |

## How expiry works

- On upload, a file gets `expiresAt = now + FILE_EXPIRY`.
- After expiry, any access returns `410 Gone` and the file is cleaned up **lazily** on that access.
- A **background sweeper** (every `CLEANUP_INTERVAL`) actively deletes expired files even if no one touches them, so nothing lingers on disk.

## Project structure

```
sharejer/
  cmd/server/main.go     # entry point: config, routes, sweeper, graceful shutdown
  internal/
    config/              # env (+ .env) configuration
    handlers/            # HTTP handlers (upload, download, meta, qr, delete, config)
    storage/             # Storage interface + local disk impl
    db/                  # SQLite setup + queries
    model/               # File, FileMeta, response structs
    sweeper/             # background expired-file cleanup goroutine
  frontend/              # Svelte 5 + Vite + Tailwind SPA (own package.json/Dockerfile)
  uploads/               # stored files (gitignored)
  sharejer.db            # SQLite file (gitignored)
  Dockerfile             # builds the Go server image
  docker-compose.yml     # api + frontend services
```

## Notes

- The upload limit is enforced via `http.MaxBytesReader`, so oversized uploads are rejected while reading rather than buffered into memory.
- Files live on local disk under `uploads/`; only metadata is stored in SQLite — never the bytes.
- Uploads stream to disk and downloads stream back out via `io.Reader`/`io.Writer`.
- Filenames are sanitized to prevent path traversal.
- Passwords are bcrypt-hashed; plaintext and hashes are never returned.
- Share links point at the frontend page `/f/{id}` (a styled download screen), not the raw JSON API.
- The server handles `SIGINT`/`SIGTERM` for graceful shutdown (drains in-flight requests, stops the sweeper).