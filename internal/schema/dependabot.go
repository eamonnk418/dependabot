package schema

type Schedule struct {
	Interval string `yaml:"interval"`
	Time     string `yaml:"time"`
}

type Update struct {
	PackageEcosystem string    `yaml:"package-ecosystem"`
	Directory        string    `yaml:"directory"`
	Schedule         *Schedule `yaml:"schedule"`
}

type Dependabot struct {
	Version int       `yaml:"version"`
	Updates []*Update `yaml:"updates"`
}
