package database

import (
	"database/sql"
	"github.com/hmaier-dev/checklist-tool/internal/structs"
	"log"

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
		ita TEXT, 
		name TEXT NOT NULL,
		ticket TEXT,
		model TEXT,
		path TEXT NOT NULL UNIQUE CHECK (length(path) == 30),
		yaml TEXT
	);
	`
	_, err = db.Exec(createStmt)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db
}

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
	insertStmt := `INSERT INTO checklists (imei, ita, name, ticket, model, path, yaml) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(insertStmt, cl.IMEI, cl.ITA, cl.Name, cl.Ticket, cl.Model, cl.Path, cl.Yaml)
	if err != nil {
		log.Fatal("Failed to insert entry: ", err)
	}

}

func PathAlreadyExists(db *sql.DB, imei string) bool {
	var exists int
	query := `SELECT COUNT(*) FROM checklists WHERE path = ?`
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
	query := `SELECT imei, ita, name, ticket, model, path, yaml FROM checklists WHERE path = ?`
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

func GetAllEntrysReversed(db *sql.DB) ([]*structs.ChecklistEntry, error) {
	query := `SELECT imei, ita, name, ticket, model, path FROM checklists ORDER BY id DESC`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error while doing '%s' the database: %s", query, err)
		return nil, err
	}
	var allEntries []*structs.ChecklistEntry
	for rows.Next() {
		var entry structs.ChecklistEntry
		if err := rows.Scan(&entry.IMEI, &entry.ITA, &entry.Name, &entry.Ticket, &entry.Model, &entry.Path); err != nil {
			log.Fatalf("Error scanning row: %s", err)
			return nil, err
		}
		allEntries = append(allEntries, &entry)
	}
	return allEntries, nil
}

func UpdateYamlByPath(db *sql.DB, path string, yamlData string) {
	_, err := db.Exec("UPDATE checklists SET yaml = ? WHERE path = ?", yamlData, path)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}

}
func DeleteEntryByPath(db *sql.DB, path string) {
	_, err := db.Exec("DELETE FROM checklists WHERE path = ?", path)
	if err != nil {
		log.Fatal("Error updating database:", err)
	}
}
