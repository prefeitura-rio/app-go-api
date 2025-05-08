FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o api ./cmd/server

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder /app/api .
COPY .env .env
ENV GIN_MODE=release
CMD ["./api"]