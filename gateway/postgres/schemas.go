package postgres

// schema intial DB setup, define tables.
var schema = `
CREATE TABLE IF NOT EXISTS person (
    id text unique NOT NULL,
    first_name text,
    last_name text,
    email text,
    timestamp int
);

CREATE TABLE IF NOT EXISTS place (
    id text unique NOT NULL,
    country text,
    city text NULL,
    comments text[] NULL,
    telcode integer,
    timestamp int
)`
