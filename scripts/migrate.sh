#!/bin/sh

set -euo pipefail

if [ "${RUN_MIGRATIONS:-true}" != "true" ]; then
	echo "[migrate] RUN_MIGRATIONS is not true, skipping migrations"
	exit 0
fi

: "${POSTGRES_USER:?missing}"
: "${POSTGRES_PASSWORD:?missing}"
: "${POSTGRES_DB:?missing}"

PGHOST="${POSTGRES_CONTAINER_HOST:-postgres_weather_container}"
PGPORT="${POSTGRES_CONTAINER_PORT:-5432}"

DSN="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${PGHOST}:${PGPORT}/${POSTGRES_DB}?sslmode=disable"

echo "[migrate] Waiting for Postgres at ${PGHOST}:${PGPORT}..."
for i in $(seq 1 30); do
	if nc -z ${PGHOST} ${PGPORT} >/dev/null 2>&1; then
		echo "[migrate] Postgres is up"
		break
	fi
	echo "[migrate] still waiting (${i}/30)"; sleep 2
	if [ "$i" -eq 30 ]; then
		echo "[migrate] timeout waiting for Postgres" >&2
		exit 1
	fi
done

# Ensure goose is available (Dockerfile.migrations installs it to PATH)
command -v goose >/dev/null 2>&1 || { echo "[migrate] goose not found" >&2; exit 1; }

echo "[migrate] Applying migrations from /app/migrations"
goose -dir /app/migrations postgres "$DSN" up

echo "[migrate] Done"
