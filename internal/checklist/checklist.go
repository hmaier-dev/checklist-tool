package checklist


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
