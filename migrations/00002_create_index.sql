-- +goose Up
CREATE INDEX idx_confirmed ON weather_subscriptions (confirmed);

-- +goose Down
DROP INDEX IF EXISTS idx_confirmed;