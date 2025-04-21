package database

import (
	"database/sql"
	"log"
	"github.com/hmaier-dev/checklist-tool/internal/structs"

	_ "github.com/mattn/go-sqlite3"
)

var DBfilePath string
var EmptyChecklist []byte
var EmptyChecklistItemsArray []*structs.ChecklistItem

// Initialize database
func Init() *sql.DB {
	db, err := sql.Open("sqlite3", DBfilePath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create the devices table if it doesn't exist
	createStmt := `
	CREATE TABLE IF NOT EXISTS templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		empty_yaml TEXT
	);
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		template_id INTEGER NOT NULL,
		data TEXT NOT NULL,
		path TEXT NOT NULL UNIQUE CHECK (length(path) == 30),
		yaml TEXT,
		FOREIGN KEY (template_id)
			REFERENCES templates (id)
	);
	CREATE TABLE IF NOT EXISTS custom_fields (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		template_id INTEGER NOT NULL,
		key TEXT NOT NULL,
		desc TEXT NOT NULL,
		FOREIGN KEY (template_id)
			REFERENCES templates (id)
	);
	`
	_, err = db.Exec(createStmt)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db
}

func NewChecklistTemplate(db *sql.DB, template_name string, file_content string, fields []string, desc []string){
	// Is there already a template with the same name?
	checkStmt := `SELECT name FROM templates WHERE name = ?`
	var exist string
	result := db.QueryRow(checkStmt, template_name)
	err := result.Scan(&exist)
	if err != sql.ErrNoRows {
		log.Printf("A template with the name '%s' is already present. \n Error: %q", template_name, err)
		return 
	}
	// Add a new entry in templates (the main table)
	newTemplateStmt := `INSERT INTO templates (name, empty_yaml) VALUES (?, ?)`
	_, err = db.Exec(newTemplateStmt, template_name, file_content)
	if err != nil{
		log.Fatalf("Couldn't insert a new template.\n Error: %q \n", err)
	}
	// Get the id of the newly 
	selectStmt := `SELECT id FROM templates WHERE name = ?`
	row := db.QueryRow(selectStmt, template_name)
	var id int
	err = row.Scan(&id)
	if err != nil {
		log.Fatal("Failed to scan entry: ", err)
		return
	}
	// Add custom input fields including description
	for i := range fields{
		newFieldStmt := `INSERT INTO custom_fields (template_id, key, desc) VALUES (?, ?, ?)`
		_, err = db.Exec(newFieldStmt, id, fields[i], desc[i])
		if err != nil{
			log.Fatalf("Error while inserting custom fields: %q\n", err)
		}
	}

}

func GetAllTemplates(db *sql.DB) []structs.ChecklistTemplate{
	selectStmt := `SELECT * FROM templates`
	rows, err := db.Query(selectStmt)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []structs.ChecklistTemplate
	for rows.Next() {
		var tmpl structs.ChecklistTemplate
		if err := rows.Scan(&tmpl.Id,&tmpl.Name,&tmpl.Empty_yaml); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, tmpl)
	}
	return all
}


func GetAllFieldsForChecklist(db *sql.DB, template_name string)[]structs.CustomFields{
	selectStmt := `SELECT cf.*
								FROM custom_fields cf
								JOIN templates t ON cf.template_id = t.id
								WHERE t.name = ?;`
	rows, err := db.Query(selectStmt, template_name)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []structs.CustomFields
	for rows.Next() {
		var fields structs.CustomFields
		if err := rows.Scan(&fields.Id, &fields.Template_id, &fields.Key, &fields.Desc); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, fields)
	}
	return all


}


// ------------------------------------------------------------------------------------
// Old Functions

func NewEntry(db *sql.DB, form structs.FormularData) {

	cl := structs.ChecklistEntry{
		IMEI:   form.IMEI,
		ITA:    form.ITA,
		Name:   form.Name,
		Ticket: form.Ticket,
		Model:  form.Model,
		Path:   form.Path,
		Yaml:   string(EmptyChecklist),
	}

	if PathAlreadyExists(db, cl.Path) == true {
		return
	}

	// Prepare the INSERT statement
	insertStmt := `INSERT INTO entries (imei, ita, name, ticket, model, path, yaml) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(insertStmt, cl.IMEI, cl.ITA, cl.Name, cl.Ticket, cl.Model, cl.Path, cl.Yaml)
	if err != nil {
		log.Fatal("Failed to insert entry: ", err)
	}

}

func PathAlreadyExists(db *sql.DB, imei string) bool {
	var exists int
	query := `SELECT COUNT(*) FROM entries WHERE path = ?`
	err := db.QueryRow(query, imei).Scan(&exists)
	if err != nil {
		log.Fatal("Failed to check if Path exists: ", err)
	}

	if exists > 0 {
		return true
	}
	return false

}

func GetDataByPath(db *sql.DB, path string) (*structs.ChecklistEntry, error) {
	query := `SELECT imei, ita, name, ticket, model, path, yaml FROM entries WHERE path = ?`
	row := db.QueryRow(query, path)
	var cl structs.ChecklistEntry

	err := row.Scan(&cl.IMEI, &cl.ITA, &cl.Name, &cl.Ticket, &cl.Model, &cl.Path, &cl.Yaml)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No entry found for concerning Path %s", path)
			return nil, nil
		}
		log.Fatal("Failed to scan entry:", err)
		return nil, err
	}
  
	return &cl, nil
}

func UpdateYamlByPath(db *sql.DB, path string, yamlData string) {
	_, err := db.Exec("UPDATE entries SET yaml = ? WHERE path = ?", yamlData, path)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}

}
func DeleteEntryByPath(db *sql.DB, path string) {
	_, err := db.Exec("DELETE FROM entries WHERE path = ?", path)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}
}

func ResetChecklistEntryByPath(db *sql.DB, path string) {
	_, err := db.Exec("UPDATE entries SET yaml = ? WHERE path = ?", string(EmptyChecklist) , path)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}
}
