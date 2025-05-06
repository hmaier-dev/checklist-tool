package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

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


var DBfilePath string

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
	CREATE TABLE IF NOT EXISTS custom_fields (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		template_id INTEGER NOT NULL,
		key TEXT NOT NULL,
		desc TEXT NOT NULL,
		FOREIGN KEY (template_id)
			REFERENCES templates (id)
	);
	CREATE TABLE IF NOT EXISTS tab_desc_schema (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		template_id INTEGER NOT NULL,
		value TEXT NOT NULL,
		FOREIGN KEY (template_id)
			REFERENCES templates (id)
	);
	CREATE TABLE IF NOT EXISTS pdf_name_schema (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		template_id INTEGER NOT NULL,
		value TEXT NOT NULL,
		FOREIGN KEY (template_id)
			REFERENCES templates (id)
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
	`
	_, err = db.Exec(createStmt)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db
}

// For storing the values inside the frontmatter of the yaml
type FrontMatter struct{
	Name string 							`yaml:"name"`
	Fields []string 					`yaml:"fields"`
	Desc []string 						`yaml:"desc"`
	Tab_desc_schema []string 	`yaml:"tab_desc_schema"`
	Pdf_name_schema []string 	`yaml:"pdf_name_schema"`
}

// Takes frontmatter and the checklist yaml to create 
// a new checklist template in all required tables
func NewChecklistTemplate(db *sql.DB, matter FrontMatter, yaml string){
	// Is there already a template with the same name?
	checkStmt := `SELECT name FROM templates WHERE name = ?`
	var exist string
	result := db.QueryRow(checkStmt, matter.Name)
	err := result.Scan(&exist)
	if err != sql.ErrNoRows {
		log.Printf("A template with the name '%s' is already present. \n Error: %q", matter.Name, err)
		return 
	}
	// Add a new entry in templates (the main table)
	newTemplateStmt := `INSERT INTO templates (name, empty_yaml) VALUES (?, ?)`
	_, err = db.Exec(newTemplateStmt, matter.Name, yaml)
	if err != nil{
		log.Fatalf("Couldn't insert a new template.\n Error: %q \n", err)
	}
	// Get the id of the newly created template
	selectStmt := `SELECT id FROM templates WHERE name = ?`
	row := db.QueryRow(selectStmt, matter.Name)
	var id int
	err = row.Scan(&id)
	if err != nil {
		log.Fatal("Failed to scan entry: ", err)
		return
	}
	// Add custom input fields including description
	for i := range matter.Fields{
		newFieldStmt := `INSERT INTO custom_fields (template_id, key, desc) VALUES (?, ?, ?)`
		_, err = db.Exec(newFieldStmt, id, matter.Fields[i], matter.Desc[i])
		if err != nil{
			log.Fatalf("Error while inserting custom fields: %q\n", err)
		}
	}
	// Add column names for browser tab description
	for t := range matter.Tab_desc_schema{
		newFieldStmt := `INSERT INTO tab_desc_schema (template_id, value) VALUES (?, ?)`
		_, err = db.Exec(newFieldStmt, id, t)
		if err != nil{
			log.Fatalf("Error while inserting column names for tab description schema: %q\n", err)
		}
	}
	// Add column names for pdf name schema
	for p := range matter.Pdf_name_schema{
		newFieldStmt := `INSERT INTO tab_desc_schema (template_id, value) VALUES (?, ?)`
		_, err = db.Exec(newFieldStmt, id, p)
		if err != nil{
			log.Fatalf("Error while inserting column names for pdf name schema: %q\n", err)
		}
	}

}

func GetAllTemplates(db *sql.DB) []ChecklistTemplate{
	selectStmt := `SELECT * FROM templates`
	rows, err := db.Query(selectStmt)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []ChecklistTemplate
	for rows.Next() {
		var tmpl ChecklistTemplate
		if err := rows.Scan(&tmpl.Id,&tmpl.Name,&tmpl.Empty_yaml); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, tmpl)
	}
	return all
}

// When a non-existent template is queried an empty slice is returned
func GetAllCustomFieldsForTemplate(db *sql.DB, template_name string)[]CustomField{
	selectStmt := `SELECT cf.*
								FROM custom_fields cf
								JOIN templates t ON cf.template_id = t.id
								WHERE t.name = ?;`
	rows, err := db.Query(selectStmt, template_name)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []CustomField
	for rows.Next() {
		var fields CustomField
		if err := rows.Scan(&fields.Id, &fields.Template_id, &fields.Key, &fields.Desc); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, fields)
	}
	return all
}

func GetAllEntriesForChecklist(db *sql.DB, template_name string)[]ChecklistEntry{
	selectStmt := `SELECT e.*
								FROM entries e
								JOIN templates t ON e.template_id = t.id
								WHERE t.name = ?;`
	rows, err := db.Query(selectStmt, template_name)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []ChecklistEntry
	for rows.Next() {
		var entry ChecklistEntry
		if err := rows.Scan(&entry.Id, &entry.Template_id, &entry.Data, &entry.Path, &entry.Yaml, &entry.Date); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, entry)
	}
	return all
}

func GetChecklistTemplateByName(db *sql.DB, template_name string) ChecklistTemplate{
	selectStmt := `SELECT id, name, empty_yaml FROM templates WHERE name = ?`
	row := db.QueryRow(selectStmt, template_name)
	var templateEntry ChecklistTemplate
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

func NewEntry(db *sql.DB, entry ChecklistEntry) sql.Result {
	insertStmt := `INSERT INTO entries (template_id, data, path, yaml, date) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(insertStmt, entry.Template_id, entry.Data, entry.Path, entry.Yaml, entry.Date)
	if err != nil {
		log.Printf("Error while inserting a new entry into 'entries'.\n Error: %q \n", err)
		return nil
	}
	return result
}

func GetEntryByPath(db *sql.DB, path string) ChecklistEntry{
	selectStmt := `SELECT id, template_id, data, path, yaml, date FROM entries WHERE path = ?`
	row := db.QueryRow(selectStmt, path)
	var singleEntry ChecklistEntry
	if err := row.Scan(&singleEntry.Id, &singleEntry.Template_id, &singleEntry.Data, &singleEntry.Path, &singleEntry.Yaml, &singleEntry.Date); err != nil {
		log.Fatalf("Error scanning row: %s. \n Query: %s", err, selectStmt)
	}
	return singleEntry
}
func GetTemplateNameByID(db *sql.DB, id int) ChecklistTemplate{
	selectStmt := `SELECT id, name, empty_yaml FROM templates WHERE id = ?`
	row := db.QueryRow(selectStmt, id)
	var tmpl ChecklistTemplate
	if err := row.Scan(&tmpl.Id, &tmpl.Name, &tmpl.Empty_yaml); err != nil {
		log.Fatalf("Error scanning row: %s. \n Query: %s", err, selectStmt)
	}
	return tmpl
}

func UpdateYamlByPath(db *sql.DB, path string, yamlData string) {
	_, err := db.Exec("UPDATE entries SET yaml = ? WHERE path = ?", yamlData, path)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}
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
