package parsing

import "strings"

// ParseINISection melakukan parsing INI sederhana untuk section tertentu.
// Mengembalikan map key->value lowercase. Return nil jika section tidak ditemukan / kosong.
func ParseINISection(content string, sectionName string) map[string]string {
	res := map[string]string{}
	var inSection bool
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		line := strings.TrimSpace(l)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(line[1 : len(line)-1])
			inSection = strings.EqualFold(section, sectionName)
			continue
		}
		if !inSection {
			continue
		}
		if idx := strings.Index(line, "="); idx != -1 {
			key := strings.TrimSpace(strings.ToLower(line[:idx]))
			val := strings.TrimSpace(line[idx+1:])
			res[key] = val
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

// ParseINIClient melakukan parsing INI sederhana untuk section [client]
func ParseINIClient(content string) map[string]string {
	return ParseINISection(content, "client")
}
