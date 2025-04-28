package reset


import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
)



func Display(w http.ResponseWriter, r *http.Request){
	wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var templates = []string{
		"reset.html",
		"nav.html",
	}
	var full = make([]string,len(templates))
	var static = filepath.Join(wd, "static")
	for i, t := range templates{
		full[i] = filepath.Join(static,t)
	}
  tmpl := template.Must(template.ParseFiles(full...))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing home template: ", err)
  }

	db := database.Init()
	all := database.GetAllTemplates(db)
	active := r.URL.Query().Get("template")
  err = tmpl.Execute(w, map[string]any{
		"Active": active,
		"Nav": handlers.UpdateNav(r),
		"Templates": all,
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }

}
func Entries(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var entries_tmpl = filepath.Join(wd, "static/resetEntries.html")
	active := r.URL.Query().Get("template")
	tmpl := template.Must(template.ParseFiles(entries_tmpl))

	db := database.Init()
	entries_raw := database.GetAllEntriesForChecklist(db, active)
	custom_fields := database.GetAllCustomFieldsForTemplate(db, active)
	entries_view := handlers.BuildEntriesView(custom_fields,entries_raw)
  err = tmpl.Execute(w, map[string]any{
		"Entries": entries_view,
  })
}
func Execute(w http.ResponseWriter, r *http.Request){
  http.Redirect(w, r, "/reset", http.StatusSeeOther)
}
