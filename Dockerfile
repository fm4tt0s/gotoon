# --- stage 1: build ---
# use the official Golang image as a builder
FROM golang:1.21-alpine AS builder

# set the working directory inside the container
WORKDIR /app

# copy the go.mod file first to leverage Docker cache
COPY go.mod ./
RUN go mod download

# copy the rest of the source code
COPY . .

# build the binary with CGO disabled for a static, portable binary
RUN CGO_ENABLED=0 GOOS=linux go build -o gotoon .

# --- stage 2: run ---
# use a minimal Alpine image for the final runtime
FROM alpine:latest

# add ca-certificates in case you eventually add HTTPS/TLS support
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# copy the binary from the builder stage
COPY --from=builder /app/gotoon .

# expose the proxy port (8080) and metrics port (2112)
EXPOSE 8080
EXPOSE 2112

# run the proxy
ENTRYPOINT ["./gotoon"]