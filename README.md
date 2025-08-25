## Development
To simplifly development on Windows and Linux [`earthly`](https://docs.earthly.dev/) is used to build the project.
```cmd
earthly +run --tag="test"
```
The outputs a image called `ghcr.io/hmaier-dev/checklist-tool:test`. Run this image by using the compose stack.
```cmd
earthly +run --tag="test" && docker compose up
```
A example compose stack is within the root of the project.
## Deployment
A `compose.yml` for the checklist-tool can look like this:
```yaml
services:
  checklist:
    image: ghcr.io/hmaier-dev/checklist-tool:latest
    container_name: "checklist-tool"
    ports:
      - 8080:8080
    restart: unless-stopped
    volumes:
      - ./sqlite.db:/root/sqlite.db
    depends_on:
      - gotenberg
  gotenberg:
    image: gotenberg/gotenberg:8
    container_name: "gotenberg"
```
### gotenberg
The gotenberg-Container is used for the creation of pdfs.
The checklist-tool can be run without gotenberg, but will throw an error when a checklist gets exported: https://github.com/hmaier-dev/checklist-tool/blob/66159b446c2180e9c846cbad91a53904368872d5/internal/pdf/pdf.go#L13

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
