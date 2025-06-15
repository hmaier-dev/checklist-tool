package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

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


type DatabaseError struct {
	Message string
	Err     error // Optionally embed the original error
}

// Implement the error interface for your custom error type
func (e *DatabaseError) Error() string {
	return e.Message
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
		name TEXT NOT NULL UNIQUE,
		empty_yaml TEXT,
		file TEXT
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
func NewChecklistTemplate(db *sql.DB, matter FrontMatter, yaml string, file string) error{
	// Is there already a template with the same name?
	checkStmt := `SELECT name FROM templates WHERE name = ?`
	var exist string
	result := db.QueryRow(checkStmt, matter.Name)
	err := result.Scan(&exist)
	if err != sql.ErrNoRows {
		msg := fmt.Sprintf("A template with the name '%s' is already present.\n", matter.Name)
		return &DatabaseError{
			Message: msg,
			Err:     err,
		}
	}
	// Add a new entry in templates (the main table)
	newTemplateStmt := `INSERT INTO templates (name, empty_yaml, file) VALUES (?, ?, ?)`
	_, err = db.Exec(newTemplateStmt, matter.Name, yaml, file)
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
		return nil
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
	for _, t := range matter.Tab_desc_schema{
		newFieldStmt := `INSERT INTO tab_desc_schema (template_id, value) VALUES (?, ?)`
		_, err = db.Exec(newFieldStmt, id, t)
		if err != nil{
			log.Fatalf("Error while inserting column names for tab description schema: %q\n", err)
		}
	}
	// Add column names for pdf name schema
	for _, p := range matter.Pdf_name_schema{
		newFieldStmt := `INSERT INTO pdf_name_schema (template_id, value) VALUES (?, ?)`
		_, err = db.Exec(newFieldStmt, id, p)
		if err != nil{
			log.Fatalf("Error while inserting column names for pdf name schema: %q\n", err)
		}
	}
	return nil
}


func UpdateChecklistTemplate(db *sql.DB, matter FrontMatter, yaml string, file string) error{
	selectId := `SELECT id FROM templates where name = ?`
	row := db.QueryRow(selectId, matter.Name)
	var id string
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		msg := fmt.Sprintf("There is no template with the name '%s'. It can't be updated.\n", matter.Name)
		return &DatabaseError{
			Message: msg,
			Err:     err,
		}
	}
	// Test if the inputs in the frontmatter are sane
	// TODO: write a test for this
	fieldl := len(matter.Fields)
	descl := len(matter.Desc)
	if  fieldl != descl {
		msg := fmt.Sprintf("Length of fields (%d) and their descriptions (%d) doesn't match.\n", fieldl, descl)
		return &DatabaseError{
			Message: msg,
			Err: nil,
		}
	}

	tx, err := db.Begin()
	// To delete all entries in the different tables and just do a new insert,
	// isn't the smartest way to update the complete template.
	// But it is the easiest I can think of,
	// and that is what I'm doing in the renew functions
	RenewCustomFieldsById(db, tx, id, matter)
	RenewTabDescSchema(db, tx, id, matter)
	RenewPdfNameSchema(db, tx, id, matter)

	// Update the yaml & file in templates where id matches
	updateTmpl := `Update templates SET empty_yaml = ?, file = ? WHERE id = ?`
	_, err = db.Exec(updateTmpl, yaml, file, id)
	if err != nil{
		tx.Rollback()
		log.Fatalf("Updating 'empty_yaml' and 'file' in 'templates' failed.\n Error: %q \n", err)
	}
	
	err = tx.Commit()
	if err != nil {
		log.Printf("Error running the commit for the update.\n Error: %q \n", err)
		return &DatabaseError{
			Message: "Error while updating the checklist template.",
			Err:     err,
		}
	}
	return nil
}


// Delete old custom fields (name & desc) and new ones
func RenewCustomFieldsById(db *sql.DB, tx *sql.Tx, id string, matter FrontMatter){
	delcf := `DELETE FROM custom_fields WHERE template_id = ?`
	_, err := db.Exec(delcf, id)
	if err != nil{
		tx.Rollback()
		log.Fatalf("Delete from custom_fields failed.\n Error: %q \n", err)
	}
	for i := range matter.Fields{
		newFieldStmt := `INSERT INTO custom_fields (template_id, key, desc) VALUES (?, ?, ?)`
		_, err = db.Exec(newFieldStmt, id, matter.Fields[i], matter.Desc[i])
		if err != nil{
			tx.Rollback()
			log.Fatalf("Error while inserting custom fields: %q\n", err)
		}
	}
}

// Delete old desc-schema entries and inserts new ones
func RenewTabDescSchema(db *sql.DB, tx *sql.Tx, id string, matter FrontMatter){
	deltd := `DELETE FROM tab_desc_schema WHERE template_id = ?`
	_, err := db.Exec(deltd, id)
	if err != nil{
		tx.Rollback()
		log.Fatalf("Delete from tab_desc_schema failed.\n Error: %q \n", err)
	}
	for _, t := range matter.Tab_desc_schema{
		newFieldStmt := `INSERT INTO tab_desc_schema (template_id, value) VALUES (?, ?)`
		_, err = db.Exec(newFieldStmt, id, t)
		if err != nil{
		 tx.Rollback()
		 log.Fatalf("Error while inserting column names for tab description schema: %q\n", err)
		}
	}
}

// Delete old pdf-schema entries and inserts new ones
func RenewPdfNameSchema(db *sql.DB, tx *sql.Tx, id string, matter FrontMatter){
	delps := `DELETE FROM pdf_name_schema WHERE template_id = ?`
	_, err := db.Exec(delps, id)
	if err != nil{
		tx.Rollback()
		log.Fatalf("Delete from pdf_name_schema failed.\n Error: %q \n", err)
	}
	for _, p := range matter.Pdf_name_schema{
		newFieldStmt := `INSERT INTO pdf_name_schema (template_id, value) VALUES (?, ?)`
		_, err = db.Exec(newFieldStmt, id, p)
		if err != nil{
			tx.Rollback()
			log.Fatalf("Error while inserting column names for pdf name schema: %q\n", err)
		}
	}

}

func UpdateYamlById(db *sql.DB, id int, yaml string){
	_, err := db.Exec("UPDATE entries SET yaml = ? WHERE id = ?", yaml, id)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}
}
func UpdateDataById(db *sql.DB, id int, data string){
	_, err := db.Exec("UPDATE entries SET data = ? WHERE id = ?", data, id)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}
}


// Row in 'templates'-table
type ChecklistTemplate struct {
	Id         int
	Name       string
	Empty_yaml string
	File string
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
		if err := rows.Scan(&tmpl.Id,&tmpl.Name,&tmpl.Empty_yaml,&tmpl.File); err != nil {
					 log.Fatalf("Error while getting all templates: %s", err)
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
								WHERE t.name = ?
								ORDER BY e.id DESC;`
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

// Does return a single template because the 'name' is UNIQUE
func GetChecklistTemplateByName(db *sql.DB, template_name string) (*ChecklistTemplate, error) {
	selectStmt := `SELECT id, name, empty_yaml FROM templates WHERE name = ?`
	row := db.QueryRow(selectStmt, template_name)
	var templateEntry ChecklistTemplate
	if err := row.Scan(&templateEntry.Id, &templateEntry.Name, &templateEntry.Empty_yaml); err != nil {
			log.Printf("Error scanning row from 'templates': %s", err)
			return nil, err
	}
	return &templateEntry, nil
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
		log.Printf(`Error while inserting a new entry into 'entries'.\n
		Error: %q \n
		Data: %q \n`, err, entry)
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
	selectStmt := `SELECT id, name, empty_yaml, file FROM templates WHERE id = ?`
	row := db.QueryRow(selectStmt, id)
	var tmpl ChecklistTemplate
	if err := row.Scan(&tmpl.Id, &tmpl.Name, &tmpl.Empty_yaml, &tmpl.File); err != nil {
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


// Single entry in 'tab_desc_schema'
type TabDescriptionSchemaEntry struct{
	Id int
	Template_id int
	Value string
}

func GetTabDescriptionsByID(db *sql.DB, template_id int) []TabDescriptionSchemaEntry{
	selectStmt := `SELECT id, template_id, value FROM tab_desc_schema  WHERE template_id = ?`
	rows, err := db.Query(selectStmt, template_id)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []TabDescriptionSchemaEntry
	for rows.Next() {
		var entry TabDescriptionSchemaEntry
		if err := rows.Scan(&entry.Id,&entry.Template_id,&entry.Value); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, entry)
	}
	return all

}

// Single entry in 'pdf_name_schema'
type PdfNamingSchemaEntry struct{
	Id int
	Template_id int
	Value string
}

func GetPdfNamingByID(db *sql.DB, template_id int) []PdfNamingSchemaEntry {
	selectStmt := `SELECT id, template_id, value FROM  pdf_name_schema WHERE template_id = ?`
	rows, err := db.Query(selectStmt, template_id)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []PdfNamingSchemaEntry
	for rows.Next() {
		var entry PdfNamingSchemaEntry
		if err := rows.Scan(&entry.Id,&entry.Template_id,&entry.Value); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, entry)
	}
	return all
}

func GetAllEntries(db *sql.DB)[]ChecklistEntry{
	selectStmt := `SELECT id, template_id, data, path, yaml, date FROM entries`
	rows, err := db.Query(selectStmt)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []ChecklistEntry
	for rows.Next() {
		var entry ChecklistEntry
		if err := rows.Scan(&entry.Id,&entry.Template_id,&entry.Data,&entry.Path,&entry.Yaml,&entry.Yaml); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, entry)
	}
	return all
}

// Row in 'entries'-table plus resolved checklist-name
type EntryPlusChecklistName struct {
	Id          int
	Data        string
	Path        string
	Yaml        string
	Date int64
	TemplateName string
}


func GetAllEntriesPlusTemplateName(db *sql.DB)[]EntryPlusChecklistName{
	selectStmt := `SELECT 
    entries.id,
		entries.data,
		entries.path,
		entries.yaml,
		entries.date,
    templates.name AS template_name
		FROM entries
		JOIN templates ON entries.template_id = templates.id
		ORDER BY entries.date DESC;
		`
	rows, err := db.Query(selectStmt)
	if err != nil{
		log.Fatalf("Error while running '%s' \n Error: %q \n", selectStmt, err)
	}
	var all []EntryPlusChecklistName
	for rows.Next() {
		var entry EntryPlusChecklistName
		if err := rows.Scan(&entry.Id,&entry.Data,&entry.Path,&entry.Yaml,&entry.Date,&entry.TemplateName); err != nil {
					 log.Fatalf("Error scanning row: %s", err)
					 return nil
		}
		all = append(all, entry)
	}
	return all

}

func DeleteEntryByPath(db *sql.DB, path string){
	deleteStmt := `DELETE FROM entries WHERE path = ?`
	_, err := db.Exec(deleteStmt,path)
	if err != nil{
		log.Fatalf("Error while running %s \n", deleteStmt)
	}
}
