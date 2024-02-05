# Stage 1: Build
# Start from the latest golang base image
FROM golang:1.23-buster as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
# CGO is enabled by default, so no need to set CGO_ENABLED=1
RUN go build -a -installsuffix cgo -o main .

# Stage 2: Setup
# Start from the base image with necessary libraries
FROM debian:buster

# Install runtime dependencies for SQLite and go-sqlite3
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8083

# Command to run the executable
CMD ["./main"]
