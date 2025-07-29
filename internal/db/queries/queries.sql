-- name: CreateTrip :one
INSERT INTO trips (user_id, name, description, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTripByID :one
SELECT * FROM trips
WHERE id = $1;

-- name: ListTripsByUserID :many
SELECT * FROM trips
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteTrip :exec
DELETE FROM trips
WHERE id = $1;

-- name: CreateLocation :one
INSERT INTO locations (name, address, site_url, notes)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLocationByID :one
SELECT * FROM locations
WHERE id = $1;

-- name: ListLocations :many
SELECT * FROM locations
ORDER BY id;

-- name: AddLocationToTrip :exec
INSERT INTO trip_locations (trip_id, location_id, position)
VALUES ($1, $2, $3);

-- name: GetTripLocations :many
SELECT l.*, tl.position FROM locations l
JOIN trip_locations tl ON l.id = tl.location_id
WHERE tl.trip_id = $1
ORDER BY tl.position;

-- name: RemoveLocationFromTrip :exec
DELETE FROM trip_locations
WHERE trip_id = $1 AND location_id = $2;

-- name: Create_Trip :one
INSERT INTO trips (user_id, name, description, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, name, description, start_date, end_date, created_at, updated_at;

-- name: Create_Location :one
INSERT INTO locations (name, address, site_url, notes)
VALUES ($1, $2, $3, $4)
RETURNING id, name, address, site_url, notes, created_at;

-- name: Create_Trip_Location :exec
INSERT INTO trip_locations (trip_id, location_id, position)
VALUES ($1, $2, $3);

