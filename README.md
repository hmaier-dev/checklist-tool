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

Use this command to run it raw without `earthly`
```bash
go build -x -v -p 4 -o ./bin/cltool main.go && tailwindcss -i ./static/base.css -o ./static/style.css && ./bin/cltool -db=sqlite.db
```
## Deployment
A example `compose.yml` can be found under the root of this project.
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
  text: "Setze diesen Schlüssel, um ein Textfeld anzuzeigen."
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
| fields | Takes a list of keys, which will store user input (need to be the same length as `desc`). E.g. `fields[2] == desc[2]` |
| desc | Takes a list of quoted strings, which function as labels for the input fields  (need to be the same length as `fields`) E.g. `desc[1] == fields[1]`|
| tab_desc_schema | Defines the browser-tab-description-schema. Use the `fields` seperated by `,`. Values will be display separated by `\|` |
| pdf_name_schema | Defines how the pdf will be named. Use the `fields` seperated by `,`. Values will be display separated by `_`. **An extra field is `date` (only available in this key)** which displays the current date when exporting in `yyyyMMdd`-format. |

### Yaml

>[!NOTE]
> Note that, the `task`-string is used as an identifier and cannot be used twice!

## Motivation
At work I'm dealing with mobile devices, whose setup require multiple steps I need to keep track of. This is not just for me but also for quality assurance.
Working with/in PDFs is tireseome in serveral ways. So I decided to write this small project, which should ease my time setup up the devices.
