## Motivation
At work I'm dealing with mobile devices, whose setup require multiple steps I need to keep track of. This is not just for me but also for quality assurance.
Working with/in PDFs is tireseome in serveral ways. So I decided to write this small project, which should ease my time setup up the devices.

## Vision
The goal of this project is to implement several things like:

- Autosave function
- Searchability by
  - full name of a person
  - IMEI
  - ticket-Number
  - device-Type (de. Modell)
- PDF-creation with a good name-schema  

## Dependencys

- `scoop install wkhtmltopdf`
- `scoop install tailwindcss`

## Roadmap

- [x] Building a UI with Golang (templates/html)
- [x] Add SQLite-Database for saving JSON-Structs
- [x] Convert html to pdf
- [x] Add styling to home.html
- [x] Add styling to checklist.html
- [ ] Optional: Different Checklists for different Jobs

