# Stage 1: Build the Go application
FROM golang:1.23 AS builder
WORKDIR /app

# CGO build tools removed - Not needed for pure Go postgres driver
# RUN apt-get update && apt-get install -y --no-install-recommends build-essential && rm -rf /var/lib/apt/lists/*

# Set environment variables for Go
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
# Ensure CGO is disabled for a pure Go build
ENV CGO_ENABLED=0

# Copy go mod and sum files first to leverage Docker cache
COPY go.mod go.sum ./
# Attempt to download modules multiple times to handle potential network issues
RUN for i in 1 2 3 4 5; do go mod download && break || sleep 5s; done

# Copy the rest of the application source code
COPY . .

# Build the Go application (Pure Go, CGO disabled)
# Removed CGO flags, build tags, and GOARCH specification
RUN go build -v -a -ldflags '-s -w' -o main . \
    && ls -l /app/main # Keep check to verify binary exists

# Stage 2: Create the final lightweight image
FROM alpine:3.19
WORKDIR /app

# libc6-compat removed - Not needed without CGO
# RUN apk --no-cache add libc6-compat

# Copy the statically linked binary from the builder stage
COPY --from=builder /app/main .
# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application (using absolute path)
CMD ["/app/main"]
