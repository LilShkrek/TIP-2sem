package env

func Get(key, fallback string) string {
	value := getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
