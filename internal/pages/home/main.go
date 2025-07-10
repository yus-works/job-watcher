package home

import (
	"html/template"
	"net/http"
)

func Register(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		tmpl.ExecuteTemplate(w, "home", nil)
	}
}
