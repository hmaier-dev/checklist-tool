package upload

import (
	"testing"

	"gopkg.in/yaml.v3"
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
	yaml.Unmarshal()

	t.Run("Update YAML", func(t *testing.T) {
							UpdateChecklistYaml(old, blank)
							// compare tt.newYaml with tt.expected
							for i, item := range tt.newYaml {
									if item.Checked != tt.expected[i].Checked {
											t.Errorf("item %d (%s): expected Checked=%v, got %v",
													i, item.Text, tt.expected[i].Checked, item.Checked)
									}
							}
					})



}
