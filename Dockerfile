# Start from the latest golang base image
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN cd cmd && go build -o /trip-plan-service main.go

# Start a new stage from scratch
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /trip-plan-service .
EXPOSE 8083
CMD ["./trip-plan-service"] 