package dependabot

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/eamonnk418/dependabot/internal/schema"
)

//go:embed template/npm.tmpl
var npmTemplate string

type NpmDependabotTemplate struct{}

func NewNpmDependabotTemplate() DependabotTemplateFactory {
	return &NpmDependabotTemplate{}
}

func (t *NpmDependabotTemplate) GenerateTemplate(schema *schema.Dependabot) (*bytes.Buffer, error) {
	tmpl, err := template.New("npm").Parse(npmTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, schema); err != nil {
		return nil, err
	}

	return &buf, nil
}
