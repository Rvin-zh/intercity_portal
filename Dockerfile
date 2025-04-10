# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder
WORKDIR /app

# CGO needs to be enabled for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev 

# Set environment variables for Go
ENV GOPROXY=https://proxy.golang.org,direct
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

# Stage 2: Create the final image using Alpine
FROM alpine:3.18
WORKDIR /app

# Install SQLite and other necessary packages
RUN apk add --no-cache sqlite ca-certificates curl tzdata

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
