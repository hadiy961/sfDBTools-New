package execution

import "strings"

// ExtractMysqldumpVersion mengambil versi mysqldump dari stderr output.
func ExtractMysqldumpVersion(stderrOutput string) string {
	if stderrOutput == "" {
		return ""
	}

	for _, line := range strings.Split(stderrOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "mysqldump") && strings.Contains(line, "Ver") {
			return line
		}
	}

	return ""
}
