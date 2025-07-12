package tmpl

import (
	"bytes"
	"html/template"
	"log"
)

func Render(
	t *template.Template,
	name string,
	data any,
) string {
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, name, data)
	if err != nil {
		log.Printf("ERROR: Executing template %s: %v", name, err)
	}
	return buf.String()
}
