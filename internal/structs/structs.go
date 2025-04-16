package structs

// Represents data needed to create a new database entry
type FormularData struct {
	IMEI   string
	ITA    string
	Name   string
	Ticket string
	Model  string
	Path   string
}

// Represents a single checklist
type ChecklistEntry struct {
	IMEI   string `yaml:"imei"`
	ITA    string `yaml:"ita"`
	Name   string `yaml:"name"`
	Ticket string `yaml:"ticket"`
	Model  string `yaml:"model"`
	Path   string `yaml:"path"`
	Yaml   string `yaml:"yaml"`
}

// Single checkpoint of the list
type ChecklistItem struct {
	Task     string           `yaml:"task"`
	Checked  bool             `yaml:"checked"`
	Children []*ChecklistItem `yaml:"children,omitempty"`
	Path     string           `yaml:"Path"`
}

type NavItem struct {
	Name string
	Path string
}
