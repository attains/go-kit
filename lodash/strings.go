package lodash

func StringEmptyOrDefault(v, defaultV string) string {
	if len(v) == 0 {
		return defaultV
	}
	return v
}
