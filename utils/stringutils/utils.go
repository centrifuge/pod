package stringutils

import "strings"

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
