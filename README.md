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
pdf_name_schema: [date,fullname,typ]
---

- task: "Auf Arbeit kommen."
  checked: false
  children:
  - task: "Kaffee trinken."
    checked: false
- task: "Tickets bearbeiten."
  checked: false
```
Fields and Desc must be the same length, otherwise a field would be without a description.
For the `tab_desc_schema` and `pdf_name_schema` all values in `fields` can be used. Non existent `fields` will just be display as empty string.

Extra fields are:

- `pdf_name_schema`:
    - `date`: displays the current date in `yyyymmdd`

## Motivation
At work I'm dealing with mobile devices, whose setup require multiple steps I need to keep track of. This is not just for me but also for quality assurance.
Working with/in PDFs is tireseome in serveral ways. So I decided to write this small project, which should ease my time setup up the devices.

## Deps

- `wkhtmltopdf`: Creating pdf-documents from html
- `tailwindcss`: utility-css framework 
