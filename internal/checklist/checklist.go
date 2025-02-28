package checklist

// To received data given in the new-formular
type FormularData struct{
  IMEI string
  Name string
  Ticket  string
  Model string
}


// Represents a single checklist
type ChecklistEntry struct {
	IMEI   string `yaml:"imei"`
	Name   string `yaml:"name"`
	Ticket string `yaml:"ticket"`
	Model  string `yaml:"model"`
	Yaml   string `yaml:"yaml"`
}

// Single checkpoint of the list
type ChecklistItem struct {
        Task     string           `json:"task"`
        Checked  bool             `json:"checked"`
        Children []*ChecklistItem `json:"children,omitempty"`
        IMEI     string           `json:"imei"`
}

