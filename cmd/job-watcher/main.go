package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func main() {
	fmt.Println("hello world")

	tmpl := template.Must(template.ParseGlob("internal/tmpl/*.html"))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "home", nil)
    })

	http.HandleFunc("/hello", hello)
	http.ListenAndServe(":8080", nil)
}
