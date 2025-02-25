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
	IMEI   string `json:"imei"`
	Name   string `json:"name"`
	Ticket string `json:"ticket"`
	Model  string `json:"model"`
	Json   string `json:"json"`
}

// Single checkpoint of the list
type ChecklistItem struct {
        Task     string           `json:"task"`
        Checked  bool             `json:"checked"`
        Children []*ChecklistItem `json:"children,omitempty"`
        IMEI     string           `json:"imei"`
}

