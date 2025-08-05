-- +goose Up
-- +goose StatementBegin
ALTER TABLE trips ADD COLUMN end_position TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trips DROP COLUMN end_position;
-- +goose StatementEnd