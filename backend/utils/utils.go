package utils

// Checks if a string is present in a slice
func StringContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
