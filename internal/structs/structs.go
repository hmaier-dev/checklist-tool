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

// Represents a single checklist entry in the database
type ChecklistEntry struct {
	Id          int
	Template_id int
	Data        string
	Path        string
	Yaml        string
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

type ChecklistTemplate struct {
	Id         int
	Name       string
	Empty_yaml string
}

type CustomFields struct {
	Id          int
	Template_id int
	Key         string
	Desc        string
}
