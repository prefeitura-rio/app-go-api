FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.23.0 && \
    go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go && \
    CGO_ENABLED=0 go build -o api ./cmd/server

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/docs ./docs
COPY internal/db/migrations internal/db/migrations
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENV GIN_MODE=release
ENV RUN_MIGRATIONS=true

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./api"]