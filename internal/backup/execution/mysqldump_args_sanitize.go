package execution

import "strings"

// IsSSLMismatchRequiredButServerNoSupport detects a common client/server mismatch:
// client requires SSL but server doesn't support it.
// Example stderr:
//
//	Got error: 2026: "TLS/SSL error: SSL is required, but the server does not support it" when trying to connect
func IsSSLMismatchRequiredButServerNoSupport(stderrOutput string) bool {
	if stderrOutput == "" {
		return false
	}
	l := strings.ToLower(stderrOutput)
	return strings.Contains(l, "tls/ssl error") &&
		strings.Contains(l, "ssl is required") &&
		strings.Contains(l, "server does not support")
}

// AddDisableSSLArgs appends a CLI option to disable SSL/TLS for mysqldump.
// We prefer '--skip-ssl' for MariaDB/MySQL client compatibility.
func AddDisableSSLArgs(args []string) ([]string, bool) {
	for _, a := range args {
		al := strings.ToLower(strings.TrimSpace(a))
		if al == "--skip-ssl" || strings.HasPrefix(al, "--ssl-mode") || strings.HasPrefix(al, "--ssl=") {
			return args, false
		}
	}
	newArgs := make([]string, 0, len(args)+1)
	newArgs = append(newArgs, args...)
	newArgs = append(newArgs, "--skip-ssl")
	return newArgs, true
}

// RemoveUnsupportedMysqldumpOption attempts to remove ONE unsupported option from args based on stderr.
// This targets common mysqldump exit status 2 cases like:
// - "unknown option '--set-gtid-purged=OFF'"
// - "unknown variable 'set-gtid-purged=OFF'"
// It never removes database names (only flags starting with '-').
func RemoveUnsupportedMysqldumpOption(args []string, stderrOutput string) ([]string, string, bool) {
	stderrLower := strings.ToLower(stderrOutput)
	if !strings.Contains(stderrLower, "unknown option") && !strings.Contains(stderrLower, "unknown variable") {
		return args, "", false
	}

	lines := strings.Split(stderrOutput, "\n")
	for _, line := range lines {
		l := strings.ToLower(line)
		if !strings.Contains(l, "unknown option") && !strings.Contains(l, "unknown variable") {
			continue
		}

		tok := extractQuotedToken(line)
		if tok == "" {
			tok = extractDashedToken(line)
		}
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}

		candidate := tok
		if !strings.HasPrefix(candidate, "-") {
			// mysqldump kadang menulis tanpa leading dashes (mis: set-gtid-purged=OFF)
			if strings.Contains(candidate, "-") || strings.Contains(candidate, "=") {
				candidate = "--" + candidate
			} else {
				continue
			}
		}

		newArgs, removed, ok := removeFlagArg(args, candidate, tok)
		if ok {
			return newArgs, removed, true
		}
	}

	return args, "", false
}

func extractQuotedToken(s string) string {
	// Prefer single quotes
	if i := strings.Index(s, "'"); i >= 0 {
		if j := strings.Index(s[i+1:], "'"); j >= 0 {
			return s[i+1 : i+1+j]
		}
	}
	// Fallback to double quotes
	if i := strings.Index(s, "\""); i >= 0 {
		if j := strings.Index(s[i+1:], "\""); j >= 0 {
			return s[i+1 : i+1+j]
		}
	}
	return ""
}

func extractDashedToken(s string) string {
	// Find first occurrence of "--" or "-" and read until whitespace.
	idx := strings.Index(s, "--")
	if idx < 0 {
		idx = strings.Index(s, "-")
	}
	if idx < 0 {
		return ""
	}

	end := idx
	for end < len(s) {
		c := s[end]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			break
		}
		end++
	}
	return s[idx:end]
}

func removeFlagArg(args []string, candidate string, rawToken string) ([]string, string, bool) {
	// Prefer exact matches on candidate.
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		match := arg == candidate || arg == rawToken
		if !match {
			// Common when stderr token omits leading dashes
			match = rawToken != "" && strings.Contains(arg, rawToken)
		}
		if !match {
			continue
		}

		removed := arg
		// If option is in two-arg form ("--opt" "value"), remove the value too.
		removeNextValue := !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-")
		out := make([]string, 0, len(args))
		out = append(out, args[:i]...)
		if removeNextValue {
			out = append(out, args[i+2:]...)
		} else {
			out = append(out, args[i+1:]...)
		}
		return out, removed, true
	}

	return args, "", false
}
