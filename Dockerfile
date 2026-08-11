# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git dan ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go.mod dan go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary aplikasi
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api

# Final Runtime Stage
FROM alpine:latest

WORKDIR /app

# Copy binary dari builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

# Expose port
EXPOSE 8080

# Run aplikasi
CMD ["./main"]
