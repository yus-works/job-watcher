package tmpl

import (
	"bytes"
	"context"
	"html/template"

	"github.com/sirupsen/logrus"
)

func Render(
	log *logrus.Entry,
	ctx context.Context,
	tl *template.Template,
	name string,
	data any,
) (string, error) {
	var buf bytes.Buffer
	err := tl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		log.WithFields(logrus.Fields{
			"name": name,
		}).WithError(err).Error(
			"execute template",
		)
		return "Failed to render 'card'", err
	}
	return buf.String(), nil
}
