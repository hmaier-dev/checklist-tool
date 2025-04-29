package structs


// Single checkpoint of the list
type ChecklistItem struct {
	Task     string           `yaml:"task"`
	Checked  bool             `yaml:"checked"`
	Children []*ChecklistItem `yaml:"children,omitempty"`
	Path     string           `yaml:"Path"`
}

