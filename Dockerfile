FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN CGO_ENABLED=0 go build -o api ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY .env /app/.env
COPY internal/db/migrations internal/db/migrations
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
RUN chmod 644 /app/.env
ENV GIN_MODE=release

# Adiciona uma variável para RUN_MIGRATIONS que pode ser sobrescrita
ENV RUN_MIGRATIONS=true

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./api"]