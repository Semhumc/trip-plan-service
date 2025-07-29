-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS trip_locations;
DROP TABLE IF EXISTS trips;
DROP TABLE IF EXISTS locations;

CREATE TABLE trips (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    site_url TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE trip_locations (
    trip_id INT REFERENCES trips(id) ON DELETE CASCADE,
    location_id INT REFERENCES locations(id) ON DELETE CASCADE,
    position INT NOT NULL,
    PRIMARY KEY (trip_id, location_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trip_locations;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS trips;
-- +goose StatementEnd
