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
) (string, error) {
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, name, data)
	if err != nil {
		log.Printf("ERROR: Executing template %s: %v", name, err)
		return "Failed to render 'card'", err
	}
	return buf.String(), nil
}
