-- +goose Up
CREATE TABLE IF NOT EXISTS weather_subscriptions (

                                     id UUID PRIMARY KEY,
                                     email TEXT NOT NULL,
                                     city TEXT NOT NULL,
                                     frequency TEXT CHECK (frequency IN ('daily', 'hourly')) NOT NULL,
                                     token TEXT UNIQUE NOT NULL,
                                     confirmed BOOLEAN DEFAULT FALSE,
                                     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                     UNIQUE(email, city)
);

-- +goose Down
DROP TABLE IF EXISTS users;