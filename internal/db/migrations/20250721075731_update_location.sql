-- +goose Up
-- +goose StatementBegin
ALTER TABLE locations DROP COLUMN latitude;
ALTER TABLE locations DROP COLUMN longitude;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE locations ADD COLUMN latitude DOUBLE PRECISION;
ALTER TABLE locations ADD COLUMN longitude DOUBLE PRECISION;
-- +goose StatementEnd
