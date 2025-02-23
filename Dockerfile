FROM golang:1.24 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY internal/ .
COPY main.go .

# Build the Go application
RUN go build -o main 

# Start from a minimal image
FROM gcr.io/distroless/base-debian12

# Set working directory
WORKDIR /root/

# Copy the compiled binary
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
