# Albion Online Guild Regear Management

A full-stack MVP for managing Albion Online guild regears, build templates, inventory, and shopping lists.

## Stack

- Frontend: React + TypeScript + Vite
- Backend: Go + Gin
- Database: SQLite via `database/sql`, with migration files designed for PostgreSQL-friendly schema evolution
- Deployment: Docker Compose

## MVP Features

- Approved guild build templates
- Build screenshot uploads stored with each template
- Regear request submission and officer review
- Inventory tracking with automatic stock decrement on approval
- Missing item detection
- Shopping list generation from approved regear shortages
- Dashboard stats, member history, audit logs, seed data
- Future-ready tables for crafting recipes and Discord notification outbox

## Quick Start

```powershell
docker compose up --build
```

Then open:

- Frontend: http://localhost:5173
- API: http://localhost:8080/api/health

## Local Development

Backend:

```powershell
cd backend
go run ./cmd/api
```

Frontend:

```powershell
cd frontend
npm.cmd install
npm.cmd run dev
```

The frontend expects the API at `http://localhost:8080` by default. Set `VITE_API_URL` to override it.

## Access

The MVP opens directly as the local admin account `Blazor`. There is no login page.

## API Overview

The local MVP does not require a token. If you add auth later, routes can still accept:

```http
Authorization: Bearer <api_token>
```

### Auth

- `POST /api/auth/login`
- `GET /api/auth/me`

### Dashboard

- `GET /api/dashboard`

### Builds

- `GET /api/builds`
- `POST /api/builds` Admin/Officer
- `PATCH /api/builds/:id` Admin/Officer
- `DELETE /api/builds/:id` Admin/Officer

### Regear Requests

- `GET /api/regears`
- `POST /api/regears`
- `PATCH /api/regears/:id/status` Admin/Officer

### Inventory

- `GET /api/inventory`
- `POST /api/inventory` Admin/Officer
- `PATCH /api/inventory/:id` Admin/Officer
- `DELETE /api/inventory/:id` Admin/Officer

### Shopping Lists

- `POST /api/shopping-lists/generate` Admin/Officer
- `GET /api/shopping-lists/latest`

### Members

- `GET /api/members/history` Admin/Officer

## Project Layout

```text
backend/
  cmd/api/
  internal/
  migrations/
frontend/
  src/
docker-compose.yml
```

## Upgrade Path to PostgreSQL

The persistence layer is isolated in `backend/internal/store`. The current schema avoids SQLite-only business logic where practical. To upgrade:

1. Add a PostgreSQL driver.
2. Replace the SQLite DSN and placeholder style in the store.
3. Run the same normalized migrations with minor type adjustments if desired.

## Discord Integration Path

The backend includes `discord_outbox` for future notification jobs. Approval alerts, shopping list posts, and officer commands can be added as async workers without changing the core regear workflow.
