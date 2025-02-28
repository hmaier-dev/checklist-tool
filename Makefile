# Build the Go binary
go-build:
	mkdir -p ./bin
	go build -o bin/checklist-tool main.go

docker-build:
	docker build -t checklist-tool .

docker-run:
	docker run -d  --name checklist-tool -v /opt/checklist-tool/sqlite.db:/root/sqlite.db -p 8181:8080 checklist-tool:latest

tailwind-build:
	tailwindcss -i ./static/base.css -o ./static/style.css

clean:
	rm ./bin/*
	rm datenbank_eins.db
	touch datenbank_eins.db


