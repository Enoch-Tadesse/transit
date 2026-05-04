-- Schema for the transit backend.
--
-- Postgres owns everything durable: identities, routes, stops, the route
-- graph, and historical trip requests. What it does NOT store is per-second
-- GPS pings or live driver sessions — that's Redis territory. See the Redis
-- design doc in documents/ for how the two databases split the workload.

BEGIN;

-- ------------------------------------------------------------------
-- Extensions
-- ------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ------------------------------------------------------------------
-- Enum types
-- ------------------------------------------------------------------
CREATE TYPE user_role AS ENUM ('passenger', 'driver', 'admin');

-- tracking_status lives here as the durable last-known state. The live
-- "is this bus currently pinging?" flag lives in Redis and is authoritative
-- while a trip is in progress. Postgres catches up when Redis expires a
-- stale session.
CREATE TYPE bus_tracking_status AS ENUM ('active', 'inactive', 'stale', 'maintenance');

CREATE TYPE route_direction AS ENUM ('forward', 'reverse');

-- ------------------------------------------------------------------
-- users
--
-- One table, role is just an enum column. Keeps auth queries simple.
-- Driver-specific stuff (bus assignments) lives in the buses table
-- rather than as nullable columns here.
-- ------------------------------------------------------------------
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role            user_role NOT NULL,
    full_name       VARCHAR(120) NOT NULL,
    email           CITEXT,
    phone_number    VARCHAR(20),
    password_hash   TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_email_or_phone_present
        CHECK (email IS NOT NULL OR phone_number IS NOT NULL)
);

-- Enable citext here — users.email depends on it, and 000_init.sql may
-- or may not have run already depending on the init order.
CREATE EXTENSION IF NOT EXISTS citext;

CREATE UNIQUE INDEX uq_users_email
    ON users (email)
    WHERE email IS NOT NULL;

CREATE UNIQUE INDEX uq_users_phone
    ON users (phone_number)
    WHERE phone_number IS NOT NULL;

CREATE INDEX idx_users_role ON users (role);

-- ------------------------------------------------------------------
-- routes
-- ------------------------------------------------------------------
CREATE TABLE routes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(20) NOT NULL,      -- 'ROUTE-A', etc.
    display_name    VARCHAR(120) NOT NULL,     -- 'Route A - Merkato to Bole'
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_routes_code UNIQUE (code)
);

CREATE INDEX idx_routes_is_active ON routes (is_active);

-- ------------------------------------------------------------------
-- stops
--
-- geom is SRID 4326 (WGS84 lat/lng) — what GPS devices emit and what
-- PostGIS ST_DWithin expects. Same CRS as Redis GEOADD, so coordinates
-- are portable between the two stores without transformation.
-- ------------------------------------------------------------------
CREATE TABLE stops (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(150) NOT NULL,
    geom            GEOMETRY(Point, 4326) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stops_geom ON stops USING GIST (geom);

-- ------------------------------------------------------------------
-- route_stops
--
-- Orders stops along a route. The (route_id, stop_order) unique
-- constraint means no two stops can claim the same position on a route.
-- The (route_id, stop_id) unique constraint stops you from listing the
-- same physical stop twice on the same route — loop routes that need
-- genuine repeats would need their own schema treatment.
-- ------------------------------------------------------------------
CREATE TABLE route_stops (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id        UUID NOT NULL REFERENCES routes (id) ON DELETE CASCADE,
    stop_id         UUID NOT NULL REFERENCES stops (id) ON DELETE RESTRICT,
    stop_order      INTEGER NOT NULL,

    CONSTRAINT chk_route_stops_order_positive CHECK (stop_order >= 0),
    CONSTRAINT uq_route_stops_route_order UNIQUE (route_id, stop_order),
    CONSTRAINT uq_route_stops_route_stop UNIQUE (route_id, stop_id)
);

CREATE INDEX idx_route_stops_route_id ON route_stops (route_id);
CREATE INDEX idx_route_stops_stop_id ON route_stops (stop_id);
CREATE INDEX idx_route_stops_route_order ON route_stops (route_id, stop_order);

-- ------------------------------------------------------------------
-- buses
--
-- Each bus belongs to one driver (users.role = 'driver'). plate_number
-- is globally unique. tracking_status is the durable record — Redis is
-- the source of truth for "is this bus actively pinging right now?"
-- ------------------------------------------------------------------
CREATE TABLE buses (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id           UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    plate_number        VARCHAR(20) NOT NULL,
    tracking_status     bus_tracking_status NOT NULL DEFAULT 'inactive',
    current_route_id    UUID REFERENCES routes (id) ON DELETE SET NULL,
    current_direction   route_direction,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_buses_plate_number UNIQUE (plate_number),

    -- If you set a route, you must also set a direction (and vice versa).
    -- Both must be NULL for an idle bus, both must be set for an active one.
    CONSTRAINT chk_buses_route_requires_direction
        CHECK (
            (current_route_id IS NULL AND current_direction IS NULL)
            OR (current_route_id IS NOT NULL AND current_direction IS NOT NULL)
        )
);

CREATE INDEX idx_buses_driver_id ON buses (driver_id);
CREATE INDEX idx_buses_current_route_id ON buses (current_route_id);
CREATE INDEX idx_buses_tracking_status ON buses (tracking_status);

-- One actively tracked bus per driver at a time. A driver can register
-- multiple buses over time (or own spares), but only one can be 'active'
-- simultaneously. Inactive/historical buses don't collide on this index.
CREATE UNIQUE INDEX uq_buses_one_active_per_driver
    ON buses (driver_id)
    WHERE tracking_status = 'active';

-- ------------------------------------------------------------------
-- trip_requests
--
-- A passenger's trip search result. resolved_plan is JSONB because the
-- shape of a plan — N legs, each with route/stop/ETA-at-query-time — is
-- a read-mostly snapshot. We don't query across individual plan fields.
-- The origin and destination geometries are stored separately so we can
-- answer questions like "where do most trip requests originate?" without
-- parsing JSONB.
-- ------------------------------------------------------------------
CREATE TABLE trip_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    passenger_id        UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    origin_geom         GEOMETRY(Point, 4326) NOT NULL,
    destination_geom    GEOMETRY(Point, 4326) NOT NULL,
    resolved_plan        JSONB,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_trip_requests_status
        CHECK (status IN ('pending', 'planned', 'no_route_found', 'expired'))
);

CREATE INDEX idx_trip_requests_passenger_id ON trip_requests (passenger_id);
CREATE INDEX idx_trip_requests_origin_geom ON trip_requests USING GIST (origin_geom);
CREATE INDEX idx_trip_requests_destination_geom ON trip_requests USING GIST (destination_geom);
CREATE INDEX idx_trip_requests_created_at ON trip_requests (created_at DESC);

COMMIT;
