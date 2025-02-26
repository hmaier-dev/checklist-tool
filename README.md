## Motivation
I am writing this project in order to get rid of the tiresome _windows-explorer-pdf-template-workflow_. Right now, there is a checlist-template which I alter in the browser and save on a Windows-Share.
There are multiple things I do not like about it:

- Working with an PDF-Editor Editor
- No Autosave-Function
- No searchability, because of bad pdf-name-schema and pdf format itself
- Moving template files through the windows filesystem

## Vision
The goal of this project is to implement several functions, which will make my work-life a lot easier:
- Searchability by
  - Name (of the Person)
  - IMEI (of the device)
  - Ticket-Number
  - Device-Type (de. Modell)
- PDF-creation with a good name-schema  

## Roadmap

- [x] Building a UI with Golang (templates/html)
- [x] Add SQLite-Database for saving JSON-Structs
- [ ] Convert html to pdf
- [ ] Add styling to home.html
- [ ] Add styling to checklist.html
- [ ] Optional: Different Checklists for different Jobs
