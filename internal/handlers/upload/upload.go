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

	"github.com/adrg/frontmatter"
	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/handlers/checklist"
	"gopkg.in/yaml.v3"
)

type UploadHandler struct{}

var _ handlers.DisplayHandler = (*UploadHandler)(nil)

type TemplatesView struct {
	Id          int
	Name        string
	Columns     string
	Description string
	Tab_Schema  string
	PDF_Schema  string
}

// Sets / and all its subroutes
func (h *UploadHandler) Routes(router *mux.Router) {
	router.HandleFunc("/upload", h.Display).Methods("Get")
	router.HandleFunc("/upload", h.Execute).Methods("POST")
	sub := router.PathPrefix("/checklist").Subrouter()
	sub.HandleFunc("/delete", h.Delete).Methods("POST")
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
func FormatToTabSchema(entries []database.TabDescriptionSchemaEntry) string {
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
func FormatToPDFSchema(entries []database.PdfNamingSchemaEntry) string {
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
	var templates = []string{
		"upload/templates/upload.html",
		"upload/templates/template.html",
		"header.html",
		"nav.html",
	}
	db := database.Init()
	template_entries := database.GetAllTemplates(db)
	var all = make([]TemplatesView, len(template_entries))
	for i, t := range template_entries {
		cols := database.GetAllCustomFieldsForTemplate(db, t.Name)
		tab := database.GetTabDescriptionsByID(db, t.Id)
		pdf := database.GetPdfNamingByID(db, t.Id)
		all[i] = TemplatesView{
			Id:          t.Id,
			Name:        t.Name,
			Columns:     FormatWithColumnsWithCommas(cols),
			Description: FormatWithDescriptionWithCommas(cols),
			Tab_Schema:  FormatToTabSchema(tab),
			PDF_Schema:  FormatToPDFSchema(pdf),
		}
	}
	tmpl := handlers.LoadTemplates(templates)
	err := tmpl.Execute(w, map[string]any{
		"Templates": all,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal("", err)
	}

}

// Runs when submit-button on / is pressed
func (h *UploadHandler) Execute(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1 << 20)
	file, header, err := r.FormFile("yaml")
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, file)
	file_contents := buf.String()
	var matter database.FrontMatter
	// splits the file into the yaml frontmatter and the rest of the file
	rest, err := frontmatter.Parse(strings.NewReader(file_contents), &matter)
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

	db := database.Init()
	err = database.NewChecklistTemplate(db, matter, string(rest), file_contents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/upload", http.StatusSeeOther)
}

func (h *UploadHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

// Handles request to /checklist/download/<template_id>
// and return the raw checklist including the frontmatter as text/yaml
func (h *UploadHandler) Download(w http.ResponseWriter, r *http.Request) {
	template_id := mux.Vars(r)["id"]
	// type conversion
	id, err := strconv.Atoi(template_id)
	if err != nil {
		http.Error(w, "Cannot get template id is not an integer.", http.StatusBadRequest)
		return
	}
	db := database.Init()
	template := database.GetTemplateNameByID(db, id)
	// Setting the header before sending the file to the browser
	w.Header().Set("Content-Type", "text/yaml")
	filename := time.Now().Format("20060102") + "_" + template.Name + ".yml"
	disposition := fmt.Sprintf("attachment; filename=%s", filename)
	w.Header().Set("Content-Disposition", disposition)
	_, err = io.Copy(w, strings.NewReader(template.File))
	if err != nil {
		log.Fatalf("Couldn't send yaml file to browser.\nError: %q \n", err)
	}
}

func (h *UploadHandler) Update(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1 << 20)
	file, header, err := r.FormFile("yaml")
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, file)
	file_contents := buf.String()
	var matter database.FrontMatter
	// splits the file into the yaml frontmatter and the rest of the file
	rest, err := frontmatter.Parse(strings.NewReader(file_contents), &matter)
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
	db := database.Init()
	defer db.Close()
	// Updating tables with new data from the frontmatter
	err = database.UpdateChecklistTemplate(db, matter, string(rest), file_contents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
	}
	entries := database.GetAllEntriesForChecklist(db, matter.Name)
	// Update dataMap for concerning entries
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
		json, err := json.Marshal(dataMap)
		database.UpdateDataById(db, e.Id, string(json))
	}

	
	// Update the checklist for all relevant entries
	// but save the state of the check-points
	for _, e := range entries {
		var oldCheck []*checklist.Item
		var blankCheck []*checklist.Item
		yaml.Unmarshal([]byte(e.Yaml), &oldCheck)
		yaml.Unmarshal(rest, &blankCheck)

		var itemsMap = make(map[string]bool)
		fillHashMap(itemsMap, oldCheck)
		adoptState(itemsMap, blankCheck)

		var newCheck []byte
		newCheck, err = yaml.Marshal(blankCheck)
		if err != nil{
			log.Fatalf("Marshaling yaml wen't wrong.")
		}
		database.UpdateYamlById(db, e.Id, string(newCheck))
	}

	// Special header for htmx
	w.Header().Set("HX-Redirect", "/upload")
	w.WriteHeader(http.StatusNoContent)
}

// dissolves the multi-level checklist-struct in a one-dimensional hashmap.
// Useful when searching for a key:value
func fillHashMap(itemMap map[string]bool, checklist []*checklist.Item) {
	for _, item := range checklist {
		itemMap[item.Task] = item.Checked
		if len(item.Children) > 0 {
			fillHashMap(itemMap, item.Children)
		}
	}
}

// adopts checklist.Item.Checked if map-key fits checklist.Item.Task
func adoptState(itemMap map[string]bool, checklist []*checklist.Item) {
	for _, item := range checklist {
		if value, ok := itemMap[item.Task]; ok {
			item.Checked = value
		}
		if len(item.Children) > 0 {
			adoptState(itemMap, item.Children)
		}
	}
}

func init() {
	handlers.RegisterHandler(&UploadHandler{})
}
