FROM golang:1.24.0-alpine as base
WORKDIR /app
RUN apk update && apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

FROM base as dev
CMD ["go", "run", "./cmd/server/main.go"]

FROM base as build
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
COPY static ./static
#RUN apk add --no-cache gcc musl-dev make swag
#RUN swag init -g cmd/api/main.go -o cmd/api/docs
RUN mkdir -p /build
RUN go build -o /build/api ./cmd/api/main.go
CMD ["/build/server"]