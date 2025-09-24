-- +goose Up
ALTER TABLE weather_subscriptions
    ADD COLUMN IF NOT EXISTS confirmed_at TIMESTAMP NULL;

-- +goose Down
ALTER TABLE weather_subscriptions
    DROP COLUMN IF EXISTS confirmed_at;


