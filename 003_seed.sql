BEGIN;

-- Stops — 13 hubs across the Addis Ababa transit grid. Coordinates are
-- SRID 4326 (WGS84 lat/lng, x=longitude, y=latitude).

INSERT INTO stops (id, name, geom) VALUES
    -- Route A spine + core transfer hubs
    ('00000000-0000-0000-0000-000000000001', 'Merkato',        ST_SetSRID(ST_MakePoint(38.7400, 9.0350), 4326)),
    ('00000000-0000-0000-0000-000000000002', 'Piazza',         ST_SetSRID(ST_MakePoint(38.7530, 9.0350), 4326)),
    ('00000000-0000-0000-0000-000000000003', 'Meskel Square',  ST_SetSRID(ST_MakePoint(38.7614, 9.0107), 4326)),
    ('00000000-0000-0000-0000-000000000004', 'Bole',           ST_SetSRID(ST_MakePoint(38.7900, 8.9950), 4326)),
    -- Route B: southern arc through Megenagna up to CMC
    ('00000000-0000-0000-0000-000000000005', 'Sarbet',         ST_SetSRID(ST_MakePoint(38.7550, 8.9900), 4326)),
    ('00000000-0000-0000-0000-000000000006', 'Megenagna',      ST_SetSRID(ST_MakePoint(38.8010, 9.0180), 4326)),
    ('00000000-0000-0000-0000-000000000007', 'CMC',            ST_SetSRID(ST_MakePoint(38.8200, 9.0300), 4326)),
    -- Route C: western bypass overlapping Route B at Meskel → Megenagna
    ('00000000-0000-0000-0000-000000000008', 'Tor Hailoch',    ST_SetSRID(ST_MakePoint(38.7180, 9.0120), 4326)),
    ('00000000-0000-0000-0000-000000000009', 'Mexico',         ST_SetSRID(ST_MakePoint(38.7460, 9.0110), 4326)),
    -- Route D: northern feeder linking Shiro Meda down to Mexico
    ('00000000-0000-0000-0000-000000000010', 'Shiro Meda',     ST_SetSRID(ST_MakePoint(38.7620, 9.0620), 4326)),
    ('00000000-0000-0000-0000-000000000011', 'Amist Kilo',     ST_SetSRID(ST_MakePoint(38.7630, 9.0430), 4326)),
    -- Route E: intentionally isolated — no shared stops with any other route
    ('00000000-0000-0000-0000-000000000012', 'Gotera',         ST_SetSRID(ST_MakePoint(38.7650, 8.9820), 4326)),
    ('00000000-0000-0000-0000-000000000013', 'Kaliti',         ST_SetSRID(ST_MakePoint(38.7670, 8.9250), 4326));

-- Routes
INSERT INTO routes (id, code, display_name, is_active) VALUES
    ('10000000-0000-0000-0000-000000000001', 'ROUTE-A', 'Route A - Merkato to Bole via Meskel', TRUE),
    ('10000000-0000-0000-0000-000000000002', 'ROUTE-B', 'Route B - Sarbet to CMC via Megenagna', TRUE),
    ('10000000-0000-0000-0000-000000000003', 'ROUTE-C', 'Route C - Tor Hailoch to Megenagna via Mexico Bypass', TRUE),
    ('10000000-0000-0000-0000-000000000004', 'ROUTE-D', 'Route D - Shiro Meda to Mexico via Piazza', TRUE),
    ('10000000-0000-0000-0000-000000000005', 'ROUTE-E', 'Route E - Gotera to Kaliti Shuttle', TRUE);

-- Route-stop graph edges (stop_order = position along the route)
--
-- A: Merkato → Piazza → Meskel Square → Bole
INSERT INTO route_stops (route_id, stop_id, stop_order) VALUES
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 0),
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', 1),
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000003', 2),
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000004', 3);

-- B: Sarbet → Meskel Square → Megenagna → CMC
INSERT INTO route_stops (route_id, stop_id, stop_order) VALUES
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000005', 0),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003', 1),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000006', 2),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000007', 3);

-- C: Tor Hailoch → Mexico → Meskel Square → Megenagna
-- Runs parallel to Route B between Meskel and Megenagna — good test for
-- multi-option path ranking (two different 3-leg routes to CMC).
INSERT INTO route_stops (route_id, stop_id, stop_order) VALUES
    ('10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000008', 0),
    ('10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000009', 1),
    ('10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000003', 2),
    ('10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000006', 3);

-- D: Shiro Meda → Amist Kilo → Piazza → Mexico
INSERT INTO route_stops (route_id, stop_id, stop_order) VALUES
    ('10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000010', 0),
    ('10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000011', 1),
    ('10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000002', 2),
    ('10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000009', 3);

-- E: Gotera → Kaliti
-- No shared stops with A/B/C/D — the path finder should return
-- "no route found" for any trip involving this line.
INSERT INTO route_stops (route_id, stop_id, stop_order) VALUES
    ('10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000012', 0),
    ('10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000013', 1);

-- Drivers + buses for manual Redis ping simulation
INSERT INTO users (id, role, full_name, email, password_hash) VALUES
    ('20000000-0000-0000-0000-000000000001', 'driver', 'Abebe Kebede', 'abebe@transit.com', 'hash'),
    ('20000000-0000-0000-0000-000000000002', 'driver', 'Aster Awoke', 'aster@transit.com', 'hash'),
    ('20000000-0000-0000-0000-000000000003', 'driver', 'Chala Chaltu', 'chala@transit.com', 'hash');

INSERT INTO buses (id, driver_id, plate_number, tracking_status, current_route_id, current_direction) VALUES
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'AA-12345', 'active', '10000000-0000-0000-0000-000000000001', 'forward'),
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'AA-67890', 'active', '10000000-0000-0000-0000-000000000002', 'forward'),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', 'AA-55555', 'active', '10000000-0000-0000-0000-000000000003', 'reverse');

COMMIT;
