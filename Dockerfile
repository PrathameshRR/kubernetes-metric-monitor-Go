FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy source code
COPY . .

# Initialize modules and download dependencies
RUN go mod download && \
    go mod tidy && \
    go mod vendor && \
    go mod verify

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-w -s" -o k-monitor .

# Final stage
FROM alpine:latest
WORKDIR /app

# Add CA certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/k-monitor .

# Create a non-root user
RUN adduser -D -g '' appuser && \
    chown appuser:appuser /app/k-monitor

USER appuser

CMD ["./k-monitor"] 