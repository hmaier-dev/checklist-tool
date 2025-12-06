VERSION 0.8

deps:
  FROM golang:1.24
  WORKDIR /src
  COPY go.mod go.sum ./
  RUN go mod download

tailwindcss:
  FROM alpine/curl
  RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0-beta.8/tailwindcss-linux-x64 && \
      chmod +x tailwindcss-linux-x64 && \
      mv tailwindcss-linux-x64 tailwindcss
  SAVE ARTIFACT ./tailwindcss

sqlc:
  FROM alpine/curl
  RUN curl -sLO https://downloads.sqlc.dev/sqlc_1.30.0_linux_amd64.tar.gz && \
      tar xvf sqlc_1.30.0_linux_amd64.tar.gz
  SAVE ARTIFACT ./sqlc

build:
  FROM +deps
  COPY +tailwindcss/tailwindcss /usr/local/bin/tailwindcss
  COPY +sqlc/sqlc /usr/local/bin/sqlc
  COPY *.sql sqlc.yml ./
  COPY *.go ./
  COPY --dir internal/ static/ ./
  RUN sqlc generate
  RUN --mount=type=cache,id=go-build-cache,target=/root/.cache/go-build \
      GOOS=linux go build -o checklist-tool main.go
  RUN tailwindcss -i ./static/base.css -o ./static/style.css
  SAVE ARTIFACT ./checklist-tool AS LOCAL ./bin/checklist-tool
  SAVE ARTIFACT ./static
  SAVE ARTIFACT ./internal

run:
  FROM debian:bookworm
  LABEL org.opencontainers.image.source = "https://github.com/hmaier-dev/checklist-tool"
  ARG tag
  WORKDIR /root
  COPY +build/static /root/static/
  COPY +build/internal /root/internal/
  COPY +build/checklist-tool .
  EXPOSE 8080
  RUN echo "You need to mount sqlite with '-v /opt/checklist-tool/sqlite:/root/sqlite.db'"
  ENTRYPOINT ["./checklist-tool", "-db=sqlite.db"]
  SAVE IMAGE --push ghcr.io/hmaier-dev/checklist-tool:$tag

