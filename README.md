# Weather-API-Application

**Genesis School Test Task**

**Author:** [Daniil Herasymenko](https://github.com/DanHerasymenko)



This application implements a weather API that allows users to subscribe to weather updates for a selected city on a daily or hourly basis.

---

## Technologies Used

- Golang 1.24.0
- PostgreSQL 17.4
- Web Framework: Gin
- DB:  PostgreSQL (pgx lib)
- Migrations: Goose
- Swagger documentation: swaggo
- Deploy: Docker + Docker Compose
- Other: SMTP, HTML + JS
---

## Run with Docker

To start the application:

1. Clone the repository:
2. Create a `.env` file in the root directory of the project. You can use the provided `.env` template below.
3. Start the application using Docker Compose:
```
docker compose up --build
```

---

## Example `.env` File

```
APP_ENV=local
APP_PORT=:8080
CONTAINER_PORT_MAPPING=8080:8080
APP_BASE_URL=http://localhost:8080

#weatherapi.com key
WEATHER_API_KEY=1234567890abcdef

#PostgreSQL
POSTGRES_CONTAINER_HOST=postgres_weather_container
POSTGRES_CONTAINER_PORT=5432
POSTGRES_LOCAL_PORT=5432
POSTGRES_USER=weather_service
POSTGRES_PASSWORD=weather_service
POSTGRES_DB=weather_service
RUN_MIGRATIONS=true

#Email
SMTP_FROM=no-reply@weather_service.com
SMTP_PASSWORD=weather_service
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
```

---

## Migrations

- Migrations run automatically if `RUN_MIGRATIONS=true`.
- To run them manually inside the container:

```
goose -dir ./migrations postgres "postgres://user:password@localhost:5432/weather_service?sslmode=disable" up
```

---

## Application URLs

- Swagger docs: http://localhost:8082/swagger/index.html
- HTML form: http://localhost:8082/static

---

## Application Logic

1. User fills out the subscription form at `/static`
2. `POST /api/subscribe` is called
3. If the subscription is new or not yet confirmed, a confirmation token is generated and emailed
4. Confirmation is handled via `GET /api/subscription/confirm/{token}`
5. Unsubscription is available via `GET /api/subscription/unsubscribe/{token}`

---

## Implemented Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET    | /api/weather?city={city} | Get current weather for a given city |
| POST   | /api/subscribe | Subscribe to weather updates |
| GET    | /api/subscription/confirm/{token} | Confirm a subscription |
| GET    | /api/subscription/unsubscribe/{token} | Unsubscribe from updates |




---