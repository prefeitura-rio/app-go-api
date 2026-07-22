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
# tzdata: base de fusos do SO. O binário também embute time/tzdata como fallback,
# mas manter o pacote deixa o fuso visível para ferramentas do container.
RUN apk add --no-cache tzdata
COPY --from=builder /app/api .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/docs ./docs
COPY internal/db/migrations internal/db/migrations
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENV GIN_MODE=release
ENV RUN_MIGRATIONS=true
# Fuso do processo. O código aplica time.Local a partir de APP_TIMEZONE
# (default America/Sao_Paulo); TZ mantém o restante do container coerente.
ENV TZ=America/Sao_Paulo
ENV APP_TIMEZONE=America/Sao_Paulo

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./api"]