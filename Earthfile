VERSION 0.8

tailwindcss:
  FROM alpine/curl
  RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0-beta.8/tailwindcss-linux-x64 && \
      chmod +x tailwindcss-linux-x64 && \
      mv tailwindcss-linux-x64 tailwindcss
  SAVE ARTIFACT ./tailwindcss

deps:
  FROM golang:1.24
  WORKDIR /src
  COPY go.mod go.sum ./
  RUN go mod download

build:
  FROM +deps
  COPY +tailwindcss/tailwindcss /usr/local/bin/tailwindcss
  COPY *.go ./
  COPY --dir internal/ static/ ./
  RUN --mount=type=cache,id=go-build-cache,target=/root/.cache/go-build \
      GOOS=linux go build -o checklist-tool main.go
  RUN tailwindcss -i ./static/base.css -o ./static/style.css
  SAVE ARTIFACT ./checklist-tool AS LOCAL ./bin/checklist-tool
  SAVE ARTIFACT ./static
  SAVE ARTIFACT ./internal

run:
  FROM ghcr.io/hmaier-dev/docker-wkhtmltopdf:v0.1
  ARG tag
  WORKDIR /root
  COPY +build/static /root/static/
  COPY +build/internal /root/internal/
  COPY +build/checklist-tool .
  EXPOSE 8080
  RUN echo "You need to mount sqlite with '-v /opt/checklist-tool/sqlite:/root/sqlite.db'"
  ENTRYPOINT ["./checklist-tool", "-db=sqlite.db"]
  SAVE IMAGE ghcr.io/hmaier-dev/checklist-tool:$tag

