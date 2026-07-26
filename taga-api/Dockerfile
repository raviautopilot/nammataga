# ==========================================
# Stage 1: Build the Go binary
# ==========================================
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Install compilation prerequisites
RUN apk add --no-cache git ca-certificates

# Copy module definitions and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Compile statically linked optimized binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o taga-api .

# ==========================================
# Stage 2: Clean execution environment
# ==========================================
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and configuration from builder
COPY --from=builder /app/taga-api .
COPY --from=builder /app/config.json .

# Expose internal API service port
EXPOSE 8080

# Execute Go application
CMD ["./taga-api"]
