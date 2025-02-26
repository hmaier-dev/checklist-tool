# Build the Go binary
go-build:
	mkdir -p ./bin
	go build -o bin/checklist-tool main.go

docker-build:
	docker build -t checklist-tool .

tailwind-build:
	tailwindcss -i ./static/base.css -o ./static/style.css


clean:
	rm ./bin/*
	rm datenbank_eins.db
	touch datenbank_eins.db


