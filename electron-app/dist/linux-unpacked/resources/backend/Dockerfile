# Build stage
FROM golang:1.23 AS builder

WORKDIR /app

# Set environment variables for Go
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies with retry logic
RUN for i in 1 2 3 4 5; do go mod download && break || sleep 5; done

# Copy source code
COPY . .

# Build the application with static linking
RUN go build -a -ldflags '-extldflags "-static"' -o main .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"]
