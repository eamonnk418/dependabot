package dependabot

import (
	"os"
	"text/template"

	"github.com/eamonnk418/dependabot/internal/constant"
)

func RenderTemplate(data any) error {
	tmpl := template.Must(template.New("dependabot").ParseFiles(constant.DependabotTemplatePath))

	if err := tmpl.ExecuteTemplate(os.Stdout, "dependabot.yml", data); err != nil {
		return err
	}

	return nil
}
