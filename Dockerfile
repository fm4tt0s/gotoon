# --- Stage 1: Build ---
# Use the official Golang image as a builder
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod file first to leverage Docker cache
COPY go.mod ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary with CGO disabled for a static, portable binary
RUN CGO_ENABLED=0 GOOS=linux go build -o gotoon .

# --- Stage 2: Run ---
# Use a minimal Alpine image for the final runtime
FROM alpine:latest

# Add ca-certificates in case you eventually add HTTPS/TLS support
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/gotoon .

# Expose the proxy port (8080) and metrics port (2112)
EXPOSE 8080
EXPOSE 2112

# Run the proxy
ENTRYPOINT ["./gotoon"]