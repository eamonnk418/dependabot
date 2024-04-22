package schema

type Update struct {
	PackageEcosystem string `yaml:"package-ecosystem"`
}

type Dependabot struct {
	Version int       `yaml:"version"`
	Updates []*Update `yaml:"updates"`
}
