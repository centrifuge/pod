package stringutils

import (
	"regexp"
	"strings"
)

// RemoveDuplicates removes duplicate strings from the slice.
// Comparision is case-insensitive
func RemoveDuplicates(strs []string) []string {
	m := make(map[string]struct{})
	var res []string
	for _, s := range strs {
		ls := strings.ToLower(strings.TrimSpace(s))
		if _, ok := m[ls]; ok {
			continue
		}

		res = append(res, s)
		m[ls] = struct{}{}
	}

	return res
}

// ContainsStringMatch returns true if regex match for given str is found
func ContainsStringMatch(match string, str string) bool {
	elem := regexp.MustCompile(match)
	found := elem.FindAllString(str, 1)
	return (len(found) > 0)
}

// ContainsStringMatchInSlice returns true if string is found as match in slice
func ContainsStringMatchInSlice(slice []string, str string) bool {
	for _, s := range slice {
		if ContainsStringMatch(s, str) {
			return true
		}
	}

	return false
}
