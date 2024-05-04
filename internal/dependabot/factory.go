package dependabot

import (
	"bytes"
	"errors"

	"github.com/eamonnk418/dependabot/internal/schema"
)

// DependabotTemplateFactory is the interface for generating Dependabot templates.
type DependabotTemplateFactory interface {
	GenerateTemplate(*schema.Dependabot) (*bytes.Buffer, error)
}

// NewDependabotTemplateFactory creates a new instance of DependabotTemplateFactory.
func NewDependabotTemplateFactory(factory Factory) DependabotTemplateFactory {
	return &DependabotTemplate{
		factory: factory,
	}
}

// DependabotTemplate represents a concrete implementation of DependabotTemplateFactory.
type DependabotTemplate struct {
	factory Factory
}

// GenerateTemplate generates the template based on the package ecosystem.
func (t *DependabotTemplate) GenerateTemplate(schema *schema.Dependabot) (*bytes.Buffer, error) {
	switch t.factory.PackageEcosystem {
	case "npm":
		return NewNpmDependabotTemplate().GenerateTemplate(schema)
	default:
		return nil, errors.New("unsupported package ecosystem: " + t.factory.PackageEcosystem)
	}
}

// Factory represents the package ecosystem and directories information.
type Factory struct {
	PackageEcosystem string
	Directories      []string
}
