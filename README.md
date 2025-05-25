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
## Deployment
A `docker-compose.yml` for the checklist-tool can look like this:
```yaml
services:
    checklist-tool:
        contaier_name: "checklist-tool"
    volumes:
        - ./sqlite.db:/root/sqlite.db
    ports:
        - "8080:8080"
```

## Checklist
This app uses `yaml` store the checklist itself and a preceding frontmatter to store all meta-data.

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
### Frontmatter
Is the place where all the meta-data is stored.

| Key | Data  |
| --- | --- |
| fields | Takes a list of keys, which will store user input (need to be the same length as `desc`) |
| desc | Takes a list of quoted strings, which function as labels for the input fields  (need to be the same length as `fields`) |
| tab_desc_schema | Defines the browser-tab-description-schema. Values will be separated by `\|` |
| pdf_name_schema | Defines how the pdf will be named. Values will be separated by `_`. **An extra field is `date`**  which displays the current date when exporting in `yyyyMMdd`-format. |

### Yaml

>[!NOTE]
> Note that, the `task`-string is used as an identifier and cannot be used twice!

## Motivation
At work I'm dealing with mobile devices, whose setup require multiple steps I need to keep track of. This is not just for me but also for quality assurance.
Working with/in PDFs is tireseome in serveral ways. So I decided to write this small project, which should ease my time setup up the devices.

## Deps
- `wkhtmltopdf`: Creating pdf-documents from html
- `tailwindcss`: utility-css framework 
