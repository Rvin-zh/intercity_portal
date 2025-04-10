# Stage 1: Build the Go application
FROM golang:1.23 AS builder
WORKDIR /app

# CGO needs to be enabled for SQLite
RUN apt-get update && apt-get install -y --no-install-recommends build-essential sqlite3 && rm -rf /var/lib/apt/lists/*

# Set environment variables for Go
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
# Enable CGO for SQLite
ENV CGO_ENABLED=1

# Copy go mod and sum files first to leverage Docker cache
COPY go.mod go.sum ./
# Attempt to download modules multiple times to handle potential network issues
RUN for i in 1 2 3 4 5; do go mod download && break || sleep 5s; done

# Copy the rest of the application source code
COPY . .

# Build the Go application (with CGO enabled for SQLite)
RUN go build -v -o main . \
    && ls -l /app/main # Keep check to verify binary exists

# Stage 2: Create the final image using Ubuntu
FROM ubuntu:22.04
WORKDIR /app

# Install SQLite and other necessary packages
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    sqlite3 \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create data directory for SQLite
RUN mkdir -p /app/data && chmod 755 /app/data

# Copy the binary from the builder stage
COPY --from=builder /app/main .
# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application (using absolute path)
CMD ["/app/main"]
