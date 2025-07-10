package router

import (
	"html/template"
	"net/http"

	"github.com/yus-works/jod-watcher/internal/pages/home"
)

func RegisterHandlers(tmpl *template.Template) {
	http.HandleFunc("/", home.Register(tmpl))
}
