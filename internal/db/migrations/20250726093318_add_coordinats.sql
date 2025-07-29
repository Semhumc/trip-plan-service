-- +goose Up
-- +goose StatementBegin
ALTER TABLE locations ADD COLUMN latitude DECIMAL(9,6);
ALTER TABLE locations ADD COLUMN longitude DECIMAL(9,6);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE locations DROP COLUMN latitude;
ALTER TABLE locations DROP COLUMN longitude;
-- +goose StatementEnd
