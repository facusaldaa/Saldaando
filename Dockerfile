FROM golang:1.23-alpine AS builder

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o botGastosPareja ./cmd/bot

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/botGastosPareja .

# Create data directory for SQLite
RUN mkdir -p /root/data

# Expose port (if needed in future)
# EXPOSE 8080

CMD ["./botGastosPareja"]

