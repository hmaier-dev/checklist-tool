package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"database/sql"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
	"github.com/gorilla/mux"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/server"
	"github.com/hmaier-dev/checklist-tool/internal/handlers/checklist"
)

type UploadHandler struct{
	Router *mux.Router
	DB *sql.DB
}

var _ handlers.DisplayHandler = (*UploadHandler)(nil)

func (h *UploadHandler) New(srv *server.Server){
	h.Router = srv.Router	
	h.DB = srv.DB
}

type TemplatesView struct {
	Id          int64
	Name        string
	Columns     string
	Description string
	Tab_Schema  string
	PDF_Schema  string
}

// Sets /upload and all its subroutes
func (h *UploadHandler) Routes() {
	h.Router.HandleFunc("/upload", h.Display).Methods("Get")
	h.Router.HandleFunc("/upload", h.Execute).Methods("POST")
	sub := h.Router.PathPrefix("/checklist").Subrouter()
	h.Router.HandleFunc("/upload/delete", h.Delete).Methods("POST")
	sub.HandleFunc(`/download/{id:\d*}`, h.Download).Methods("GET")
	sub.HandleFunc("/update", h.Update).Methods("POST")
}

// Uses CustomField.Key
func FormatWithColumnsWithCommas(fields []database.CustomField) string {
	var result string
	for i, c := range fields {
		if i == len(fields)-1 {
			result += c.Key
		} else {
			result += c.Key + ","
		}
	}
	return result
}

// Uses CustomField.Desc
func FormatWithDescriptionWithCommas(fields []database.CustomField) string {
	var result string
	for i, c := range fields {
		if i == len(fields)-1 {
			result += "'" + c.Desc + "'"
		} else {
			result += "'" + c.Desc + "',"
		}
	}
	return result
}

// Format to Tab Schema
func FormatToTabSchema(entries []database.TabDescSchema) string {
	var result string
	for i, t := range entries {
		if i == len(entries)-1 {
			result += t.Value
		} else {
			result += t.Value + " | "
		}
	}
	return result
}

// Format to PDF Schema
func FormatToPDFSchema(entries []database.PdfNameSchema) string {
	var result string
	for i, t := range entries {
		if i == len(entries)-1 {
			result += t.Value + ".pdf"
		} else {
			result += t.Value + "_"
		}
	}
	return result
}

// Return html to http.ResponseWriter for /
func (h *UploadHandler) Display(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var templates = []string{
		"upload/templates/upload.html",
		"upload/templates/template.html",
		"header.html",
		"nav.html",
	}
	q := database.New(h.DB)
	allTemplates, err := q.GetAllTemplates(ctx)
	if err != nil{
		msg := "Couldn't fetch all templates from database."
		log.Println(msg)
		http.Error(w,msg,http.StatusInternalServerError)
	}
	var all = make([]TemplatesView, len(allTemplates))
	for i, t := range allTemplates {
		cols, err := q.GetCustomFieldsByTemplateName(ctx, t.Name)
		tab, err := q.GetTabDescriptionsByTemplateID(ctx, t.ID)
		pdf, err := q.GetPdfNamingByTemplateID(ctx, t.ID)
		if err != nil{
			msg := fmt.Sprintf("Error while fetching the database. \n Error: %v\n", err)
			log.Println(msg)
			http.Error(w,msg,http.StatusInternalServerError)
		}
		all[i] = TemplatesView{
			Id:          t.ID,
			Name:        t.Name,
			Columns:     FormatWithColumnsWithCommas(cols),
			Description: FormatWithDescriptionWithCommas(cols),
			Tab_Schema:  FormatToTabSchema(tab),
			PDF_Schema:  FormatToPDFSchema(pdf),
		}
	}
	tmpl := handlers.LoadTemplates(templates)
	err = tmpl.Execute(w, map[string]any{
		"Templates": all,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

type FrontMatter struct{
	Name string 							`yaml:"name"`
	Fields []string 					`yaml:"fields"`
	Desc []string 						`yaml:"desc"`
	Tab_desc_schema []string 	`yaml:"tab_desc_schema"`
	Pdf_name_schema []string 	`yaml:"pdf_name_schema"`
}

// Runs when submit-button is pressed
// Actual upload happens here!
func (h *UploadHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseMultipartForm(1 << 20)
	file, header, err := r.FormFile("yaml")
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, file)
	fileContents := buf.String()
	var matter FrontMatter
	// splits the file into the yaml frontmatter and the rest of the file
	rest, err := frontmatter.Parse(strings.NewReader(fileContents), &matter)
	if err != nil {
		http.Error(w, "Error while parsing frontmatter", http.StatusBadRequest)
		log.Printf("Error while parsing frontmatter.\n %q\n", err)
		return
	}
	var result any
	err = yaml.Unmarshal([]byte(rest), &result)
	if err != nil {
		log.Fatalf("Error while validating the yaml in %s: %q\n", header.Filename, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// I'm gonna do several exec-queries. Afterwards they are gonna be used TOGETHER in the same context.
	// If one fails, all changes should be rolled back. That way the data keeps consitent.
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil{
		http.Error(w,"Database error.",http.StatusInternalServerError)
	}
	defer func(){
		if err != nil{
			rbErr := tx.Rollback()
			if rbErr != nil{
				log.Printf("Rollback error: %v", rbErr)	
			}
		}
	}()
	qtx := database.New(h.DB).WithTx(tx)
	
	id, err := qtx.InsertNewChecklistTemplate(ctx, database.InsertNewChecklistTemplateParams{
		Name: matter.Name,
		EmptyYaml: sql.NullString{String: string(rest), Valid: true},
		File: sql.NullString{String: fileContents, Valid: true},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
		// Add custom input fields including description
	for i := range matter.Fields{
		arg := database.InsertCustomFieldParams{
			TemplateID: id,
			Key: matter.Fields[i],
			Desc: matter.Desc[i],
		}
		err := qtx.InsertCustomField(ctx, arg)
		if err != nil{
			msg := fmt.Sprintf("Error while inserting frontmatter values into 'custom_fields'. \n Error: %v\n", err)
			log.Println(msg)
			http.Error(w,msg,http.StatusInternalServerError)
			return
		}
	}

	// These column names tell the application later which,
	// which value from database.Entry.Data (json-slice) should be displayed

	// Add column names for browser tab description.
	for _, t := range matter.Tab_desc_schema{
		arg := database.InsertTabDescSchemaParams{
			TemplateID: id,
			Value: t,
		}
		err := qtx.InsertTabDescSchema(ctx, arg)
		if err != nil{
			msg := fmt.Sprintf("Error while inserting frontmatter values into 'tab_desc_schema'. \n Error: %v\n", err)
			log.Println(msg)
			http.Error(w,msg,http.StatusInternalServerError)
			return
		}
	}
	// Add column names for pdf name schema
	for _, p := range matter.Pdf_name_schema{
		arg := database.InsertPdfNameSchemaParams{
			TemplateID: id,
			Value: p,
		}
		err := qtx.InsertPdfNameSchema(ctx, arg)
		if err != nil{
			msg := fmt.Sprintf("Error while inserting frontmatter values into 'tab_desc_schema'. \n Error: %v\n", err)
			log.Println(msg)
			http.Error(w,msg,http.StatusInternalServerError)
			return
		}
	}
	
	tx.Commit()
	http.Redirect(w, r, "/upload", http.StatusSeeOther)
}

func (h *UploadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.FormValue("id")
	// For deleting we need 'id' as int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		panic(err)
	}

	// When deleting a checklist, all relics should be deleted.
	// Reason is, to keep the db-tables small and clean.
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil{
		http.Error(w,"Database error.",http.StatusInternalServerError)
	}
	defer func(){
		if err != nil{
			rbErr := tx.Rollback()
			if rbErr != nil{
				log.Printf("Rollback error: %v", rbErr)	
			}
		}
	}()
	qtx := database.New(h.DB).WithTx(tx)
	qtx.DeleteTemplateByID(ctx, id)

	// Clean all meta-data tables
	qtx.DeleteCustomFieldsByTemplateID(ctx, id)
	qtx.DeleteTabDescSchemaByTemplateID(ctx, id)
	qtx.DeletePdfNameSchemaByTemplateID(ctx, id)
	// Remove all entries from the active list
	qtx.DeleteEntriesByTemplateID(ctx, id)
	// Template for the checklist itself. It won't be able for selection.
	qtx.DeleteTemplateByID(ctx, id)

	err = tx.Commit()
	if err != nil {
		log.Printf("Error while deleting a template, with alls its relics.\n Error: %q \n", err)
		http.Error(w, "Error while deleting.", http.StatusInternalServerError)
	}
	// Special header for htmx
	w.Header().Set("HX-Redirect", "/upload")
	w.WriteHeader(http.StatusNoContent)
}

// Handles request to /checklist/download/<templateId>
// and return the raw checklist including the frontmatter as text/yaml
func (h *UploadHandler) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	templateIdStr := mux.Vars(r)["id"]
	// database needs int64
	templateId, err := strconv.ParseInt(templateIdStr, 10, 64)
	if err != nil {
		http.Error(w, "'id' is weird.", http.StatusBadRequest)
		return
	}
	q := database.New(h.DB)
	template, err := q.GetTemplateById(ctx, templateId)
	// Setting the header before sending the file to the browser
	w.Header().Set("Content-Type", "text/yaml")
	filename := time.Now().Format("20060102") + "_" + template.Name + ".yml"
	disposition := fmt.Sprintf("attachment; filename=%s", filename)
	w.Header().Set("Content-Disposition", disposition)
	var f string
	if template.File.Valid{
		f = template.File.String
	}else{
		http.Error(w, "Couldn't download file. It is NULL.", http.StatusBadRequest)
		return
	}
	_, err = io.Copy(w, strings.NewReader(f))
	if err != nil {
		msg := "Couldn't send yaml file to browser."
		log.Printf("%s\nError: %q \n", msg, err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
}

func (h *UploadHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseMultipartForm(1 << 20)
	file, header, err := r.FormFile("yaml")
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, file)
	fileContents := buf.String()
	var matter FrontMatter
	// splits the file into the yaml frontmatter and the rest of the file
	rest, err := frontmatter.Parse(strings.NewReader(fileContents), &matter)
	if err != nil {
		http.Error(w, "Error while parsing frontmatter", http.StatusBadRequest)
		log.Printf("Error while parsing frontmatter.\n %q\n", err)
		return
	}
	var result any
	err = yaml.Unmarshal([]byte(rest), &result)
	if err != nil {
		log.Printf("Error while validating the yaml in %s: %q\n", header.Filename, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tx, err := h.DB.Begin()
	if err != nil{
		http.Error(w,"Database error.",http.StatusInternalServerError)
	}
	defer func(){
		if err != nil{
			rbErr := tx.Rollback()
			if rbErr != nil{
				log.Printf("Rollback error: %v", rbErr)	
			}
		}
	}()
	qtx := database.New(h.DB).WithTx(tx)

	// The templateName gets declared in the frontmatter
	id, err := qtx.GetTemplateIdByName(ctx, matter.Name)
	if err == sql.ErrNoRows{
		msg := fmt.Sprintf("No template with the name '%s' exists. I can't get updated.", matter.Name)
		http.Error(w, msg, http.StatusBadRequest)
	}
	fl := len(matter.Fields)
	dl := len(matter.Desc)
	if fl != dl{
		msg := "Amount of fields and descriptions are uneven."
		http.Error(w, msg, http.StatusBadRequest)
	}

	// The following code should be as functions, but am very lazy...

	qtx.DeleteCustomFieldsByTemplateID(ctx, id)
	for i := range matter.Fields{
		arg := database.InsertCustomFieldParams{
			TemplateID: id,
			Key: matter.Fields[i],
			Desc: matter.Desc[i],
		}
		err := qtx.InsertCustomField(ctx, arg)
		if err != nil{
			msg := fmt.Sprintf("Error while inserting frontmatter values into 'custom_fields'. \n Error: %v\n", err)
			log.Println(msg)
			http.Error(w,msg,http.StatusInternalServerError)
			return
		}
	}

	// Updating the tab-desc-schema by deleting and inserting
	qtx.DeleteTabDescSchemaByTemplateID(ctx, id)
	for _, t := range matter.Tab_desc_schema{
		arg := database.InsertTabDescSchemaParams{
			TemplateID: id,
			Value: t,
		}
		err := qtx.InsertTabDescSchema(ctx, arg)
		if err != nil{
			msg := fmt.Sprintf("Error while inserting frontmatter values into 'tab_desc_schema'. \n Error: %v\n", err)
			log.Println(msg)
			http.Error(w,msg,http.StatusInternalServerError)
			return
		}
	}
	
	// Updating the pdf-schema by deleting and inserting
	qtx.DeletePdfNameSchemaByTemplateID(ctx, id)
	for _, p := range matter.Pdf_name_schema{
		arg := database.InsertPdfNameSchemaParams{
			TemplateID: id,
			Value: p,
		}
		err := qtx.InsertPdfNameSchema(ctx, arg)
		if err != nil{
			msg := fmt.Sprintf("Error while inserting frontmatter values into 'tab_desc_schema'. \n Error: %v\n", err)
			log.Println(msg)
			http.Error(w,msg,http.StatusInternalServerError)
			return
		}
	}
	
	arg := database.UpdateTemplateByIdParams{
		EmptyYaml: sql.NullString{String: string(rest), Valid: true},
		File: sql.NullString{String: fileContents, Valid: true},
		ID: id,
	}
	qtx.UpdateTemplateById(ctx, arg)

	// After updating the checklist-template itself and all concerning meta-data tables,
	// now we update the already existing entries for this template
	entries, err := qtx.GetEntriesByTemplateName(ctx, matter.Name)

	if err != nil{
		msg := fmt.Sprintf("Couldn't return entries for template: '%s'.\n Error: %v\n", matter.Name, err)
		log.Println(msg)
		http.Error(w,msg,http.StatusInternalServerError)
		return
	}
	// Update dataMap ('entry.Data' json-arry) for all concerning entries
	for _, e := range entries {
		var dataMap map[string]string
		err := json.Unmarshal([]byte(e.Data), &dataMap)
		if err != nil {
			log.Fatal("Error while unmarshaling json into map[string]string.")
		}
		for _, c := range matter.Fields {
			if _, exists := dataMap[c]; !exists {
				dataMap[c] = ""
			}
		}
		j, err := json.Marshal(dataMap)
		arg := database.UpdateDataByIdParams{
			Data: string(j),
			ID: e.ID,
		}
		qtx.UpdateDataById(ctx, arg)

	}
	
	// Update the checklist for all relevant entries
	// but save the state of the check-points
	for _, e := range entries {
		var oldCheck []*checklist.Item
		var blankCheck []*checklist.Item
		var y string
		if e.Yaml.Valid{
			y = e.Yaml.String
		}else{
			log.Println("'yaml'-field in database was NULL.")
			return
		}
		yaml.Unmarshal([]byte(y), &oldCheck)
		yaml.Unmarshal(rest, &blankCheck)

		var itemsMap = make(map[string]*checklist.Item)
		//
		fillHashMap(itemsMap, oldCheck)
		adoptState(itemsMap, blankCheck)

		var newCheck []byte
		newCheck, err = yaml.Marshal(blankCheck)
		if err != nil{
			log.Fatalf("Marshaling yaml wen't wrong.")
		}
		arg := database.UpdateYamlByIdParams{
			Yaml: sql.NullString{Valid: true, String: string(newCheck)},
			ID: e.ID,
		}
		qtx.UpdateYamlById(ctx, arg)
	}

	tx.Commit()
	// Special header for htmx
	w.Header().Set("HX-Redirect", "/upload")
	w.WriteHeader(http.StatusNoContent)
}

// dissolves the multi-level checklist-struct in a one-dimensional hashmap.
// Useful when searching for a key:value
func fillHashMap(itemMap map[string]*checklist.Item, items []*checklist.Item) {
	for _, item := range items {
		val := checklist.Item{
			Checked: item.Checked,
			Text: item.Text,
		}
		itemMap[item.Task] = &val
		if len(item.Children) > 0 {
			fillHashMap(itemMap, item.Children)
		}
	}
}

// adopts checklist.Item.Checked if map-key fits checklist.Item.Task
func adoptState(itemMap map[string]*checklist.Item, checklist []*checklist.Item) {
	for _, item := range checklist {
		if value, ok := itemMap[item.Task]; ok {
			item.Checked = value.Checked
			// Reversed logic
			// When the new value is not nil, but the old value was nil
			// Otherwise, don't overwritte the existing value
			if value.Text != nil && item.Text == nil{
				item.Text = value.Text
			}
		}
		if len(item.Children) > 0 {
			adoptState(itemMap, item.Children)
		}
	}
}

func init() {
	handlers.RegisterHandler(&UploadHandler{})
}
