package tmpl

import (
	"bytes"
	"context"
	"html/template"

	"github.com/yus-works/job-watcher/internal/logging"
)

func Render(
	ctx context.Context,
	t *template.Template,
	name string,
	data any,
) (string, error) {
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, name, data)
	if err != nil {
		logging.From(ctx).Error(
			"Failed to execute template",
			"name", name,
			"err", err,
		)
		return "Failed to render 'card'", err
	}
	return buf.String(), nil
}
