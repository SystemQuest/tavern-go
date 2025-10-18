# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY server.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Runtime stage
FROM alpine:latest

# Install wget for healthcheck
RUN apk add --no-cache wget

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/server .

# Expose port
EXPOSE 5555

# Run the server
CMD ["./server"]
