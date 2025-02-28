FROM golang:1.24 AS builder

# Set the working directory
WORKDIR /app

RUN touch sqlite.db

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o checklist-tool main.go

RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0-beta.8/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && \
    mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

RUN tailwindcss -i ./static/base.css -o ./static/style.css

FROM debian:bookworm

# Set working directory
WORKDIR /root/

RUN apt-get update && apt-get install -y wkhtmltopdf

COPY --from=builder /app/static/ ./static/
COPY --from=builder /app/checklist_allgemein.json .
COPY --from=builder /app/checklist-tool .

EXPOSE 8080

# You need to mount sqlite with '-v /opt/checklist-tool/sqlite:/root/sqlite.db'
CMD /root/checklist-tool -db=/root/sqlite.db -json=/root/checklist_allgemein.json
