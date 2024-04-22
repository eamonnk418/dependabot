package types

type Schedule struct {
	Interval string `yaml:"interval"`
}

type Update struct {
	PackageEcosystem string   `yaml:"package-ecosystem"`
	Directory        string   `yaml:"directory"`
	Schedule         Schedule `yaml:"schedule"`
}

type DependabotConfiguration struct {
	Version int      `yaml:"version"`
	Updates []Update `yaml:"updates"`
}
