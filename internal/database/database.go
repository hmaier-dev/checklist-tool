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
	CREATE TABLE IF NOT EXISTS checklists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		imei TEXT NOT NULL,
		name TEXT,
		ticket TEXT,
		model TEXT,
		yaml TEXT
	);
	`
	_, err = db.Exec(createStmt)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
  return db
}

func NewEntry(db *sql.DB, form structs.FormularData){

  cl := structs.ChecklistEntry{
		IMEI:   form.IMEI ,
		Name:   form.Name,
		Ticket: form.Ticket,
		Model:  form.Model,
		Yaml:   string(EmptyChecklist),
	}

  if IMEIalreadyExists(db, cl.IMEI) == true{
    return
  }

  // Prepare the INSERT statement
	insertStmt := `INSERT INTO checklists (imei, name, ticket, model, yaml) VALUES (?, ?, ?, ?, ?)`
  _, err := db.Exec(insertStmt, cl.IMEI, cl.Name, cl.Ticket, cl.Model, cl.Yaml)
	if err != nil {
		log.Fatal("Failed to insert entry: ", err)
	}

}

func IMEIalreadyExists(db *sql.DB, imei string)(bool){
	var exists int
	query := `SELECT COUNT(*) FROM checklists WHERE imei = ?`
  err := db.QueryRow(query, imei).Scan(&exists)
	if err != nil {
		log.Fatal("Failed to check if IMEI exists:", err)
	}

	if exists > 0 {
    return true
	}
  return false

 
}

func GetDataByIMEI(db *sql.DB, imei string)(*structs.ChecklistEntry, error){
	query := `SELECT imei, name, ticket, model, yaml FROM checklists WHERE imei = ?`
	row := db.QueryRow(query, imei)
	var cl structs.ChecklistEntry

	err := row.Scan(&cl.IMEI, &cl.Name, &cl.Ticket, &cl.Model, &cl.Yaml)
	if err != nil {
		if err == sql.ErrNoRows {
			// No entry found with the given IMEI
			log.Printf("No entry found for IMEI %s", imei)
			return nil, nil
		}
		log.Fatal("Failed to scan entry:", err)
		return nil, err
	}

	return &cl, nil
}

func GetAllEntrysReversed(db *sql.DB)([]*structs.ChecklistEntry, error){
  query := `SELECT imei, name, ticket, model FROM checklists ORDER BY id DESC`
  rows, err := db.Query(query)
  if err != nil {
    log.Fatalf("Error while doing '%s' the database: %s", query, err)
    return nil, err
  }
  var allEntries []*structs.ChecklistEntry
  for rows.Next(){
   var entry structs.ChecklistEntry
    if err := rows.Scan(&entry.IMEI, &entry.Name, &entry.Ticket, &entry.Model); err != nil {
        log.Fatalf("Error scanning row: %s", err)
        return nil, err
    }
    allEntries = append(allEntries, &entry)
  }
  return allEntries, nil
}

func UpdateYamlByIMEI(db *sql.DB, imei string, yamlData string){
  _, err := db.Exec("UPDATE checklists SET yaml = ? WHERE imei = ?", yamlData, imei)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}

}
