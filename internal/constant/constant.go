package constant

func SupportedDependabotPackageEcosystems() map[string][]string {
	return map[string][]string{
		"npm": {"package.json"},
	}
}
