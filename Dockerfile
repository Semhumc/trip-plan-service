# Start from the latest golang base image
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY . .

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Build the main application
RUN cd cmd && go build -o /trip-plan-service main.go

# Start a new stage from scratch
FROM alpine:latest
WORKDIR /root/

# Copy goose binary from builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy the main application
COPY --from=builder /trip-plan-service .

# Copy migration files
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

EXPOSE 8083
CMD ["./trip-plan-service"]