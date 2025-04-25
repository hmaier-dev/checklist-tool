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
		path TEXT NOT NULL UNIQUE,
		yaml TEXT,
		date INT,
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

// When a non-existent template is queried an empty slice is returned
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

func GetAllEntriesForChecklist(db *sql.DB, template_name string)[]structs.ChecklistEntry{
	selectStmt := `SELECT e.*
								FROM entries e
								JOIN templates t ON e.template_id = t.id
								WHERE t.name = ?;`
	rows, err := db.Query(selectStmt, template_name)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []structs.ChecklistEntry
	for rows.Next() {
		var entry structs.ChecklistEntry
		if err := rows.Scan(&entry.Id, &entry.Template_id, &entry.Data, &entry.Path, &entry.Yaml, &entry.Date); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, entry)
	}
	return all
}

func GetChecklistTemplateByName(db *sql.DB, template_name string) structs.ChecklistTemplate{
	selectStmt := `SELECT id, name, empty_yaml FROM templates WHERE name = ?`
	row := db.QueryRow(selectStmt, template_name)
	var templateEntry structs.ChecklistTemplate
	if err := row.Scan(&templateEntry.Id, &templateEntry.Name, &templateEntry.Empty_yaml); err != nil {
			log.Fatalf("Error scanning row: %s", err)
	}
	return templateEntry
}

func DoesPathAlreadyExisit(db *sql.DB, path string)bool{
	checkStmt := `SELECT path FROM entries WHERE path = ?`
	var exist string
	err := db.QueryRow(checkStmt, path).Scan(&exist)
	if err == sql.ErrNoRows{
		return false
	}
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	return true
}

func NewEntry(db *sql.DB, entry structs.ChecklistEntry) sql.Result {
	insertStmt := `INSERT INTO entries (template_id, data, path, yaml, date) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(insertStmt, entry.Template_id, entry.Data, entry.Path, entry.Yaml, entry.Date)
	if err != nil {
		log.Printf("Error while doing '%s'.\n Error: %q \n", insertStmt ,err)
		return nil
	}
	return result
}

// ------------------------------------------------------------------------------------
// Old Functions

// func NewEntry(db *sql.DB, form structs.FormularData) {

// 	cl := structs.ChecklistEntry{
// 		IMEI:   form.IMEI,
// 		ITA:    form.ITA,
// 		Name:   form.Name,
// 		Ticket: form.Ticket,
// 		Model:  form.Model,
// 		Path:   form.Path,
// 		Yaml:   string(EmptyChecklist),
// 	}

// 	if PathAlreadyExists(db, cl.Path) == true {
// 		return
// 	}

// 	// Prepare the INSERT statement
// 	insertStmt := `INSERT INTO entries (imei, ita, name, ticket, model, path, yaml) VALUES (?, ?, ?, ?, ?, ?, ?)`
// 	_, err := db.Exec(insertStmt, cl.IMEI, cl.ITA, cl.Name, cl.Ticket, cl.Model, cl.Path, cl.Yaml)
// 	if err != nil {
// 		log.Fatal("Failed to insert entry: ", err)
// 	}

// }

// func PathAlreadyExists(db *sql.DB, imei string) bool {
// 	var exists int
// 	query := `SELECT COUNT(*) FROM entries WHERE path = ?`
// 	err := db.QueryRow(query, imei).Scan(&exists)
// 	if err != nil {
// 		log.Fatal("Failed to check if Path exists: ", err)
// 	}

// 	if exists > 0 {
// 		return true
// 	}
// 	return false

// }

// func GetDataByPath(db *sql.DB, path string) (*structs.ChecklistEntry, error) {
// 	query := `SELECT imei, ita, name, ticket, model, path, yaml FROM entries WHERE path = ?`
// 	row := db.QueryRow(query, path)
// 	var cl structs.ChecklistEntry

// 	err := row.Scan(&cl.IMEI, &cl.ITA, &cl.Name, &cl.Ticket, &cl.Model, &cl.Path, &cl.Yaml)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			log.Printf("No entry found for concerning Path %s", path)
// 			return nil, nil
// 		}
// 		log.Fatal("Failed to scan entry:", err)
// 		return nil, err
// 	}
//   
// 	return &cl, nil
// }

// func UpdateYamlByPath(db *sql.DB, path string, yamlData string) {
// 	_, err := db.Exec("UPDATE entries SET yaml = ? WHERE path = ?", yamlData, path)
// 	if err != nil {
// 		log.Fatal("Error updating database:", err)
// 	}

// }
// func DeleteEntryByPath(db *sql.DB, path string) {
// 	_, err := db.Exec("DELETE FROM entries WHERE path = ?", path)
// 	if err != nil {
// 		log.Fatal("Error updating database:", err)
// 	}
// }

// func ResetChecklistEntryByPath(db *sql.DB, path string) {
// 	_, err := db.Exec("UPDATE entries SET yaml = ? WHERE path = ?", string(EmptyChecklist) , path)
// 	if err != nil {
// 		log.Fatal("Error updating database:", err)
// 	}
// }
