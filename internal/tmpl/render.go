package tmpl

import (
	"bytes"
	"html/template"
)

func Render(
	t *template.Template,
	name string,
	data any,
) string {
	var buf bytes.Buffer
	_ = t.ExecuteTemplate(&buf, name, data)
	return buf.String()
}
