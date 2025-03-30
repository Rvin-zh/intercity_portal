FROM golang:1.20

# Set DNS servers to use Google's public DNS
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf && \
    echo "nameserver 8.8.4.4" >> /etc/resolv.conf

# Set environment variables to disable network dependency during build
ENV GOPROXY=off
ENV GOSUMDB=off
ENV CGO_ENABLED=0

WORKDIR /app

# Copy the entire application
COPY . .

# Build the application using the vendored dependencies
RUN go build -mod=vendor -o main .

# Expose port 5000
EXPOSE 5000

# Command to run the executable
CMD ["./main"]