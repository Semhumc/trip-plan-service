-- trips.sql

-- name: CreateTrip :one
-- DÜZELTİLDİ: "finish_position" -> "end_position" olarak değiştirildi.
INSERT INTO trips (user_id, name, description, start_date, end_date, start_position, end_position)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, name, description, start_date, end_date, start_position, end_position, created_at, updated_at;

-- name: GetTripByID :one
-- DÜZELTİLDİ: "finish_position" -> "end_position" olarak değiştirildi.
SELECT id, user_id, name, description, start_date, end_date, start_position, end_position, created_at, updated_at
FROM trips
WHERE id = $1;

-- name: ListTripsByUserID :many
-- DÜZELTİLDİ: "finish_position" -> "end_position" olarak değiştirildi.
SELECT id, user_id, name, description, start_date, end_date, start_position, end_position, created_at, updated_at
FROM trips
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteTrip :exec
DELETE FROM trips
WHERE id = $1;


-- locations.sql

-- name: CreateLocation :one
-- GÜNCELLENDİ: latitude ve longitude eklendi. Parametre sayıları arttı ($4 -> $6).
INSERT INTO locations (name, address, site_url, notes, latitude, longitude)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, address, site_url, notes, latitude, longitude, created_at;

-- name: GetLocationByID :one
-- GÜNCELLENDİ: "*" yerine tüm kolonlar açıkça yazılarak yeni kolonlar eklendi.
SELECT id, name, address, site_url, notes, latitude, longitude, created_at
FROM locations
WHERE id = $1;

-- name: ListLocations :many
-- GÜNCELLENDİ: "*" yerine tüm kolonlar açıkça yazılarak yeni kolonlar eklendi.
SELECT id, name, address, site_url, notes, latitude, longitude, created_at
FROM locations
ORDER BY id;


-- trip_locations.sql (İlişkisel Sorgular)

-- name: AddLocationToTrip :exec
INSERT INTO trip_locations (trip_id, location_id, position)
VALUES ($1, $2, $3);

-- name: GetTripLocations :many
-- GÜNCELLENDİ: "l.*" yerine tüm location kolonları açıkça yazılarak yeni kolonlar eklendi.
SELECT l.id, l.name, l.address, l.site_url, l.notes, l.latitude, l.longitude, l.created_at, tl.position
FROM locations l
JOIN trip_locations tl ON l.id = tl.location_id
WHERE tl.trip_id = $1
ORDER BY tl.position;

-- name: RemoveLocationFromTrip :exec
DELETE FROM trip_locations
WHERE trip_id = $1 AND location_id = $2;