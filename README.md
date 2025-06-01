# Weather-API-Application



**Author:** [Daniil Herasymenko](https://github.com/DanHerasymenko)


This application implements a Weather API that allows users to subscribe to daily or hourly weather updates for a selected city. Subscribed users receive periodic emails with forecasts, and the system automatically schedules updates based on the selected frequency. The application also supports subscription confirmation, unsubscription, and background processing of updates.

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
2. Create a `.env` file in the root directory of the project. You can use the provided `.env` template below. (be sure to set the `WEATHER_API_KEY` and check default `8080` and `5432` ports to be free)
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



1. User fills out the subscription form at `/static`.

2. `POST /api/subscribe` is called:
    - If the subscription is **new or not confirmed**, a unique confirmation token is generated and sent via email.
    - If already confirmed: 409 - email already subscribed.

3. User confirms the subscription via `GET /api/subscription/confirm/{token}`:
    - The confirmation activates the subscription and schedules automatic weather updates.

4. Periodic update logic:
    - Based on the selected frequency (`daily` or `hourly`), a background scheduler starts sending weather updates.
    - Each confirmed subscription runs in its own background routine.
   
5. User can unsubscribe anytime via `GET /api/subscription/unsubscribe/{token}`:
    - This action stops future updates and removes the subscription.
    
---

## Implemented Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET    | /api/weather?city={city} | Get current weather for a given city |
| POST   | /api/subscribe | Subscribe to weather updates |
| GET    | /api/subscription/confirm/{token} | Confirm a subscription |
| GET    | /api/subscription/unsubscribe/{token} | Unsubscribe from updates |


---

## Example Email Output

Once a weather update is triggered, subscribers receive an email like the following:

```
Subject: Irpin forecast

Weather for Irpin:
- temperature: 15.8°C
- humidity: 52%
- description: Patchy rain nearby
```
