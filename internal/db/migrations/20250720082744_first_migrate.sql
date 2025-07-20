-- +goose Up
-- +goose StatementBegin

CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    site_url TEXT,
);

CREATE TABLE trips (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_point_id INT REFERENCES locations(id),
    end_point_id INT REFERENCES locations(id),
    nights INT,
    single_or_multiple BOOLEAN,
    camp_or_caravan BOOLEAN,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    max_participants INT,
    budget DOUBLE PRECISION,
    is_favorite BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE trip_locations (
    trip_id INT REFERENCES trips(id) ON DELETE CASCADE,
    location_id INT REFERENCES locations(id) ON DELETE CASCADE,
    position INT,
    PRIMARY KEY (trip_id, location_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trip_locations;
DROP TABLE IF EXISTS trips;
DROP TABLE IF EXISTS locations;
-- +goose StatementEnd
