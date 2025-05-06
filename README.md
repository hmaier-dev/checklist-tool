## Development
To reach reproducibility [`earthly`](https://docs.earthly.dev/) is used to build the project. With 
```cmd
earthly +run --tag="test"
```
you will output a image called `ghcr.io/hmaier-dev/checklist-tool:test` which is exposed on port 8080. Run it like this:
```cmd
# Windows
docker run --rm --name checklist-tool -v $pwd\sqlite.db:/root/sqlite.db -p 8080:8080 ghcr.io/hmaier-dev/checklist-tool:test
# Linux
docker run --rm --name checklist-tool -v $PWD/sqlite.db:/root/sqlite.db -p 8080:8080 ghcr.io/hmaier-dev/checklist-tool:test
```

## Checklist Example
A checklist can look like this:
```yaml
---
name: auf arbeit gehen
fields: [fullname,ticket,typ]
desc: ["Nachname, Vorname","Ticket Number","Modell"]
tab_desc_schema: [fullname,typ]
pdf_name_schema: [ticket,fullname,typ]
---

- task: "Auf Arbeit kommen."
  checked: false
  children:
  - task: "Kaffee trinken."
    checked: false
- task: "Tickets bearbeiten."
  checked: false
```


## Motivation
At work I'm dealing with mobile devices, whose setup require multiple steps I need to keep track of. This is not just for me but also for quality assurance.
Working with/in PDFs is tireseome in serveral ways. So I decided to write this small project, which should ease my time setup up the devices.

## Deps

- `wkhtmltopdf`: Creating pdf-documents from html
- `tailwindcss`: utility-css framework 

## Roadmap

- [x] Building a UI with Golang (templates/html)
- [x] Add SQLite-Database for saving JSON-Structs
- [x] Convert html to pdf
- [x] Add styling to home.html
- [x] Add styling to checklist.html
- [ ] Optional: Different Checklists for different Jobs

