package utils

// StringLengthEqual returns true if length of the string == n
func StringLengthEqual(msg string, n int) bool {
	if len(msg) != n {
		return false
	}

	return true
}

// StringEmpty returns true if the string is empty
func StringEmpty(msg string) bool {
	return StringLengthEqual(msg, 0)
}
