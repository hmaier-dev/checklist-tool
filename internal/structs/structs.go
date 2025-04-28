package structs


// Single checkpoint of the list
type ChecklistItem struct {
	Task     string           `yaml:"task"`
	Checked  bool             `yaml:"checked"`
	Children []*ChecklistItem `yaml:"children,omitempty"`
	Path     string           `yaml:"Path"`
}

// Row in 'templates'-table
type ChecklistTemplate struct {
	Id         int
	Name       string
	Empty_yaml string
}

// Row in 'custom_fields'-table
type CustomField struct {
	Id          int
	Template_id int
	Key         string
	Desc        string
}

// Row in 'entries'-table
type ChecklistEntry struct {
	Id          int
	Template_id int
	Data        string
	Path        string
	Yaml        string
	Date int64
}

