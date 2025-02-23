# Build the Go binary
go-build:
	mkdir -p ./bin
	go build -o bin/checklist-tool main.go

docker-build:
	docker build -t checklist-tool .
