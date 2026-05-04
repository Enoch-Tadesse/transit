# 🚌 Transit — Real-Time Bus Tracking & Trip Planning Backend

A backend service that matches passengers to the fastest bus route — including multi-transfer journeys — using live driver GPS data, a PostGIS-backed reference graph, and a Redis-driven real-time layer.

Built to answer one question fast and reliably: **"Which bus, and how long until it gets to me?"**

---

## Why this project

Most transit-tracking demos stop at "show a bus on a map." This one is designed the way a production system would need to be:

- **Durable data and volatile data are deliberately separated.** Routes, stops, and bus identity live in Postgres. GPS pings, driver sessions, and live position live in Redis — because writing dozens of location pings per second straight into a relational database doesn't scale, and it doesn't need to.
- **Staleness is handled by expiry, not polling.** A driver who stops pinging is detected via Redis TTL expiry and keyspace notifications — not a cron job scanning every session every few seconds.
- **Writes are atomic where it matters.** A single Lua script updates session state, the geo index, and the raw location payload together, so a crash mid-ping can never leave the system in a half-updated, inconsistent state.
- **The data model plans for transfers, not just single-bus trips.** Stops and routes form a graph specifically so multi-leg journeys can be resolved, ranked, and returned — not bolted on later.

---

## Architecture

```
┌─────────────┐        ┌──────────────────────┐        ┌─────────────────────┐
│  Passenger  │──────▶ │  Matching / Routing  │◀────── │  Driver (mobile)    │
│  (trip req) │        │       Engine         │        │   (location ping)   │
└─────────────┘        └──────────┬───────────┘        └──────────┬──────────┘
                                  │                               │
                    ┌─────────────┴──────────────────┐            │
                    ▼                                ▼            ▼
         ┌──────────────────────┐         ┌────────────────────────────────┐
         │      PostgreSQL      │         │             Redis              │
         │      + PostGIS       │         │                                │
         │  routes · stops ·    │         │  driver_session (HASH, 60s TTL)│
         │  route_stops · buses │         │  geo:active_drivers (GEO)      │
         │  · trip_requests     │         │  location:{bus_id} (JSON, 60s) │
         └──────────────────────┘         └────────────────────────────────┘
```

Postgres is the system of record for anything that must survive a restart. Redis is the system of record for "is this bus alive right now" — nothing more, nothing less.

---

## Core features

| Capability | How it works |
|---|---|
| **Multi-leg trip planning** | `stops` and `route_stops` form a graph; the matching engine walks it to find single- and multi-bus paths between origin and destination, ranked by ETA and transfer count. |
| **Nearest-stop resolution** | PostGIS `GEOGRAPHY`/`GEOMETRY` + GiST indexes power `ST_DWithin`/KNN queries so passengers are routed to the correct stop *on their chosen route*, not just the closest stop overall. |
| **Live driver tracking** | Drivers ping every 5–10s; `GEOADD` into a Redis geo index enables sub-second `GEOSEARCH` for "which active buses are near this stop." |
| **Automatic staleness detection** | Each driver session is a Redis hash with a 60s TTL, refreshed on every ping. Expiry triggers a keyspace-notification sweeper that removes the bus from the geo index and demotes its status in Postgres — no polling required. |
| **Atomic ping ingestion** | A single Lua script performs a check-and-write: verify the session is still active, update the hash, refresh the raw location string, and update the geo index — all in one round trip, so partial writes are impossible. |
| **One active bus per driver** | Enforced at the database level with a partial unique index (`WHERE tracking_status = 'active'`), not just application logic. |

---

## Tech stack

- **Database:** PostgreSQL 15+ with PostGIS (spatial queries, GiST indexes)
- **Real-time layer:** Redis (Hashes, Geo/Sorted Sets, keyspace notifications, Lua scripting)
- **Data integrity:** UUID PKs, `CITEXT` for case-insensitive email lookups, enum types for role/status fields, check constraints enforcing business invariants at the schema level
- **Design docs:** PRD written in INVEST-format user stories with explicit acceptance criteria; every schema/Redis decision is documented inline with the trade-off it was chosen over

---

## Data model highlights

- `users` — single table with role segregation via enum (`passenger` / `driver` / `admin`) rather than table-per-role inheritance, keeping auth queries simple.
- `buses` — belongs to a driver; a partial unique index guarantees only one *actively tracked* bus per driver at a time, while preserving full history of past buses.
- `route_stops` — junction table with a compound unique constraint on `(route_id, stop_order)`, guaranteeing a strict, gap-free sequence per route.
- `trip_requests` — stores the resolved leg plan as `JSONB` (a read-mostly snapshot) while keeping raw origin/destination as PostGIS points for future spatial analytics.

---

## Real-time layer highlights

- **`transit:driver_session:{bus_id}`** — HASH, 60s TTL, refreshed per ping. The single source of truth for "is this driver currently online."
- **`transit:geo:active_drivers`** — GEO index for `GEOSEARCH`-based proximity queries; since Redis GEO has no native per-member TTL, a keyspace-notification sweeper removes stale entries on session expiry.
- **`transit:location:{bus_id}`** — short-TTL JSON string carrying speed/heading/accuracy metadata that doesn't fit cleanly into a geo index, used as the input to ETA calculation.

---

## Getting started

```bash
# 1. Spin up Postgres with PostGIS and Redis
docker run -d --name transit-postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgis/postgis:15-3.4
docker run -d --name transit-redis -p 6379:6379 redis:7

# 2. Run migrations in order
psql -h localhost -U postgres -f 000_init.sql
psql -h localhost -U postgres -f 001_schema.sql
psql -h localhost -U postgres -f 003_seed.sql

# 3. Enable Redis keyspace notifications for expiry-driven staleness detection
redis-cli CONFIG SET notify-keyspace-events Ex
```
---
