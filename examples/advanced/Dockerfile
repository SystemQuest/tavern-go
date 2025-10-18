# Build stage
FROM golang:1.23-alpine AS builder

# Install gcc for CGO (needed for sqlite3)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o server .

# Runtime stage
FROM alpine:latest

# Install sqlite runtime libraries and wget for healthcheck
RUN apk add --no-cache sqlite-libs wget

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/server .

# Expose port
EXPOSE 5000

# Run the server
CMD ["./server"]
