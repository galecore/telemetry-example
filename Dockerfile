# Stage 1: Build the Go binaries
FROM golang:1.22-alpine as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the httpserver binary
WORKDIR /app/cmd/httpserver
RUN CGO_ENABLED=0 GOOS=linux go build -o httpserver

# Build the httpclient binary
WORKDIR /app/cmd/httpclient
RUN CGO_ENABLED=0 GOOS=linux go build -o httpclient

# Stage 2: Copy the binaries to a new image
FROM alpine:latest

WORKDIR /app

# Copy the binaries from builder
COPY --from=builder /app/cmd/httpserver/httpserver .
COPY --from=builder /app/cmd/httpclient/httpclient .

# Expose port 8080 for the httpserver
EXPOSE 8080

# Run the httpserver by default
CMD ["./httpserver"]
