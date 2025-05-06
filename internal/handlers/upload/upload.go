package upload

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
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
	sub := router.PathPrefix("/upload").Subrouter()
	sub.HandleFunc("", h.Display).Methods("Get")
	sub.HandleFunc("", h.Execute).Methods("POST")
}

// Uses CustomField.Key
func FormatWithColumnsWithCommas(fields []database.CustomField) string{
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
func FormatWithDescriptionWithCommas(fields []database.CustomField) string{
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
func FormatToTabSchema(entries []database.TabDescriptionSchemaEntry) string{
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
func FormatToPDFSchema(entries []database.PdfNamingSchemaEntry) string{
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
		"nav.html",
		"upload/templates/template.html",
	}
	db := database.Init()
	template_entries := database.GetAllTemplates(db)
	var all = make([]TemplatesView, len(template_entries))
	for i, t := range template_entries {
		cols := database.GetAllCustomFieldsForTemplate(db, t.Name)
		tab := database.GetTabDescriptionsByID(db, t.Id)
		pdf := database.GetPdfNamingByID(db, t.Id)
		all[i] = TemplatesView{
			Id:   t.Id,
			Name: t.Name,
			Columns: FormatWithColumnsWithCommas(cols),
			Description: FormatWithDescriptionWithCommas(cols),
			Tab_Schema: FormatToTabSchema(tab),
			PDF_Schema: FormatToPDFSchema(pdf),
		}
	}
	tmpl := handlers.LoadTemplates(templates)
	err := tmpl.Execute(w, map[string]any{
		"Templates": all,
		"Nav":       handlers.UpdateNav(r),
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
	}

	db := database.Init()
	err = database.NewChecklistTemplate(db, matter, string(rest), file_contents)
	if err != nil{
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}

	http.Redirect(w, r, "/upload", http.StatusSeeOther)
}

func init() {
	handlers.RegisterHandler(&UploadHandler{})
}
