# syntax=docker/dockerfile:1

FROM golang:alpine AS builder

WORKDIR /app

# Install git for downloading private/public modules when needed
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a small static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/bin/task-management-api ./cmd/api

FROM alpine:latest AS runner

WORKDIR /app

# Optional: certificates for outbound HTTPS calls
RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /app/bin/task-management-api /app/task-management-api
COPY --from=builder /app/web /app/web
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/task-management-api"]
