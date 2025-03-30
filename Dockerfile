FROM golang:1.20-alpine

# Install git for dependency downloads
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Update module path for Docker environment
RUN go mod init go-transportation-portal

# Build the application
RUN go build -o main .

# Expose port 5000
EXPOSE 5000

# Command to run the executable
CMD ["./main"]