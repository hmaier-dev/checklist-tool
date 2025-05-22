package upload

import (
	"testing"
)

func TestUpdateChecklistYaml(t *testing.T) {
	old := `
	- task: "Auf Arbeit kommen."
  checked: true
  children:
		- task: "Kaffee trinken."
			checked: false
		- task: "Müsli essen."
			checked: true
	- task: "Tickets bearbeiten."
		checked: false
		children:
			- task: "Kommentare schreiben."
				checked: true
	`
	blank := `
	- task: "Auf Arbeit kommen."
		checked: false
		children:
			- task: "Kaffee trinken."
				checked: false
			- task: "Müsli essen."
				checked: false
			- task: "Rauchen gehen."
				checked: false
	- task: "Tickets bearbeiten."
		checked: false
		children:
			- task: "Kommentare schreiben."
				checked: false
			- task: "Emails schreiben."
				checked: false
	`
	expected := `
	- task: "Auf Arbeit kommen."
		checked: true
		children:
			- task: "Kaffee trinken."
				checked: false
			- task: "Müsli essen."
				checked: true
			- task: "Rauchen gehen."
				checked: false
	- task: "Tickets bearbeiten."
		checked: false
		children:
			- task: "Kommentare schreiben."
				checked: true
			- task: "Emails schreiben."
				checked: false
	`




}
