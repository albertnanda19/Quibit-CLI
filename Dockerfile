# Stage 1 — Builder
FROM golang:1.25.5-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o quibit .

# Stage 2 — Runtime
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S quibit && \
    adduser -u 1001 -S quibit -G quibit

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/quibit .

# Change ownership to non-root user
RUN chown quibit:quibit /app/quibit

# Switch to non-root user
USER quibit

# Set entrypoint
ENTRYPOINT ["./quibit"]
