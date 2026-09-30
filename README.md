# Geo Treasure

Bury a treasure in real life, pin its exact GPS location, and let other hunters
find it on a shared map.

- **Frontend:** Vue 3 + Vite + Leaflet (OpenStreetMap tiles — 100% free, no API key). Deploys to Vercel.
- **Backend:** Go (chi router) + JWT auth. Runs on your own server.
- **Database:** PostgreSQL with plain `lat`/`lng` columns. Nearby search uses a
  bounding-box pre-filter plus an exact Haversine distance check.

```
geo-treasure/
├── backend/     Go API server
└── frontend/    Vue single-page app
```

## Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 14+ (running locally or remotely)

## 1. Set up PostgreSQL (no Docker)

On macOS with Homebrew:

```bash
brew install postgresql@16
brew services start postgresql@16

# Create the database and user used by the default connection string.
psql postgres <<'SQL'
CREATE ROLE geotreasure WITH LOGIN PASSWORD 'geotreasure';
CREATE DATABASE geotreasure OWNER geotreasure;
SQL
```

The backend creates its tables automatically on startup (migrations are embedded),
so there is nothing else to run.

## 2. Run the backend

```bash
cd backend
cp .env.example .env        # then edit values as needed
go run ./cmd/server
```

The API listens on `http://localhost:8080` by default.

Key environment variables (see `backend/.env.example`):

| Variable        | Purpose                                             |
| --------------- | --------------------------------------------------- |
| `PORT`          | HTTP port (default `8080`)                          |
| `DATABASE_URL`  | Postgres connection string                          |
| `JWT_SECRET`    | Secret used to sign auth tokens (use a long random) |
| `CORS_ORIGINS`  | Comma-separated allowed frontend origins            |

## 3. Run the frontend

```bash
cd frontend
cp .env.example .env        # set VITE_API_URL if backend isn't on :8080
npm install
npm run dev
```

Open `http://localhost:5173`.

## API overview

| Method | Path                        | Auth | Description                              |
| ------ | --------------------------- | ---- | ---------------------------------------- |
| POST   | `/api/auth/register`        | no   | Create account, returns token            |
| POST   | `/api/auth/login`           | no   | Log in, returns token                    |
| GET    | `/api/auth/me`              | yes  | Current user                             |
| GET    | `/api/treasures`            | opt  | List treasures. Query: `q`, `lat`,`lng`,`radius` |
| POST   | `/api/treasures`            | yes  | Bury a treasure                          |
| POST   | `/api/treasures/{id}/find`  | yes  | Mark a treasure as found                 |
| DELETE | `/api/treasures/{id}`       | yes  | Delete your own treasure                 |

Nearby example: `/api/treasures?lat=40.71&lng=-74.0&radius=2000` (radius in meters).

## Deploying

- **Frontend → Vercel:** Set the project root to `frontend/`. Vercel auto-detects
  Vite. Add an env var `VITE_API_URL` pointing to your server's public API URL.
  `vercel.json` already rewrites all routes to `index.html` for the SPA router.
- **Backend → your server:** Build a static binary and run it behind HTTPS:
  ```bash
  cd backend
  CGO_ENABLED=0 go build -o geotreasure ./cmd/server
  ```
  Set `CORS_ORIGINS` to include your Vercel domain, and use a strong `JWT_SECRET`.

## Notes on geolocation

Browsers only allow the GPS/geolocation API over `https://` (or `localhost`).
When deployed, make sure the frontend is served over HTTPS or "Use my current
location" will be blocked.
