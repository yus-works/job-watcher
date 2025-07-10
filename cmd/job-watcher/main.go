package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/yus-works/jod-watcher/internal/router"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/",
		http.StripPrefix("/static/", fs),
	)

	tmpl := template.Must(template.ParseGlob("internal/tmpl/*.html"))
	router.RegisterHandlers(tmpl)

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
