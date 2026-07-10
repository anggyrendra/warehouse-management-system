# ---- Build stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary
COPY --from=builder /app/bin/server /app/server

# Copy migrations so they can be applied on startup
COPY --from=builder /app/migrations /app/migrations

# Copy docs so /docs serves the Swagger UI
COPY --from=builder /app/docs /app/docs

EXPOSE 8080

ENTRYPOINT ["/app/server"]
