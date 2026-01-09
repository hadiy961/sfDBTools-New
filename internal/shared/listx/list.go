package listx

import "strings"

// ListTrimNonEmpty returns a new slice with each element trimmed and empty entries removed.
func ListTrimNonEmpty(list []string) []string {
	cleaned := make([]string, 0, len(list))
	for _, s := range list {
		trim := strings.TrimSpace(s)
		if trim != "" {
			cleaned = append(cleaned, trim)
		}
	}
	return cleaned
}

// StringSliceContainsFold returns true if list contains item (case-insensitive compare).
func StringSliceContainsFold(list []string, item string) bool {
	for _, v := range list {
		if strings.EqualFold(v, item) {
			return true
		}
	}
	return false
}

// CSVToCleanList converts a comma-separated string into a trimmed, non-empty slice.
func CSVToCleanList(csv string) []string {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	return ListTrimNonEmpty(parts)
}

// ListUnique returns a new slice with duplicate values removed (case-insensitive), keeping original order.
func ListUnique(list []string) []string {
	seen := make(map[string]struct{}, len(list))
	res := make([]string, 0, len(list))
	for _, v := range list {
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		res = append(res, strings.TrimSpace(v))
	}
	return res
}

// ListSubtract returns items in a that are not present in b (case-insensitive compare).
func ListSubtract(a, b []string) []string {
	if len(a) == 0 {
		return nil
	}
	if len(b) == 0 {
		return ListUnique(a)
	}
	bm := make(map[string]struct{}, len(b))
	for _, v := range b {
		bm[strings.ToLower(strings.TrimSpace(v))] = struct{}{}
	}
	out := make([]string, 0, len(a))
	seen := make(map[string]struct{}, len(a))
	for _, v := range a {
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		if _, blocked := bm[key]; !blocked {
			out = append(out, strings.TrimSpace(v))
			seen[key] = struct{}{}
		}
	}
	return out
}
