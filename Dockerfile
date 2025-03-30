FROM golang:1.20

# No need to install git, it's already in the debian-based image

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# No need to reinitialize module as we've already copied go.mod

# Build the application
RUN go build -o main .

# Expose port 5000
EXPOSE 5000

# Command to run the executable
CMD ["./main"]