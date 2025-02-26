FROM golang:1.24 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o checklist-tool main.go

RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0-beta.8/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && \
    mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

RUN tailwindcss -i ./static/base.css -o ./static/style.css

# Start from a minimal image
FROM gcr.io/distroless/base-debian12

# Set working directory
WORKDIR /root/

COPY --from=builder /app/sqlite.db .
COPY --from=builder /app/static .
COPY --from=builder /app/test_checklist.json .
COPY --from=builder /app/checklist-tool .

EXPOSE 8080
CMD ["./checklist-tool"]
