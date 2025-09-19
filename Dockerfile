FROM golang:1.24.0-alpine as base
WORKDIR /app
RUN apk update && apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

# Optional: install swag for swagger generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

FROM base as build
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
COPY static ./static

# Optional: generate swagger during build (uncomment if needed)
# RUN swag init -g cmd/api/main.go -o cmd/api/docs

RUN mkdir -p /build
RUN go build -o /build/server ./cmd/api/main.go
CMD ["/build/server"]

FROM base as dev
CMD ["go", "run", "./cmd/api/main.go"]