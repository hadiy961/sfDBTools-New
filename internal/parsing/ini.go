package parsing

import "strings"

// ParseINIClient melakukan parsing INI sederhana untuk section [client]
func ParseINIClient(content string) map[string]string {
	// Hasil parsing
	res := map[string]string{}
	var inClient bool
	// Split konten menjadi baris
	lines := strings.Split(content, "\n")
	// Iterasi setiap baris
	for _, l := range lines {
		// Bersihkan baris
		line := strings.TrimSpace(l)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		// [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(line[1 : len(line)-1])
			inClient = strings.EqualFold(section, "client")
			continue
		}
		// key=value hanya diproses jika di dalam section [client]
		if !inClient {
			continue
		}
		// key = value
		if idx := strings.Index(line, "="); idx != -1 {
			key := strings.TrimSpace(strings.ToLower(line[:idx]))
			val := strings.TrimSpace(line[idx+1:])
			res[key] = val
		}
	}
	// Jika tidak ada data, kembalikan nil
	if len(res) == 0 {
		return nil
	}
	return res
}
