package usersgrants

import (
	"sort"
	"strings"
)

func escapeMySQLLiteral(s string) string {
	// Minimal escaping for building SQL string literals safely.
	// MySQL string literal uses backslash escaping.
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

func quoteMySQLLiteral(s string) string {
	return "'" + escapeMySQLLiteral(s) + "'"
}

func quoteMySQLAccount(user, host string) string {
	return quoteMySQLLiteral(user) + "@" + quoteMySQLLiteral(host)
}

type systemUserSpec struct {
	user string
	host string // empty = any host
}

func parseSystemUserSpec(spec string) (systemUserSpec, bool) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return systemUserSpec{}, false
	}
	parts := strings.SplitN(spec, "@", 2)
	u := strings.TrimSpace(parts[0])
	if u == "" {
		return systemUserSpec{}, false
	}
	if len(parts) == 1 {
		return systemUserSpec{user: u, host: ""}, true
	}
	h := strings.TrimSpace(parts[1])
	return systemUserSpec{user: u, host: h}, true
}

func isSystemUser(account UserAccount, specs []systemUserSpec) bool {
	u := strings.ToLower(strings.TrimSpace(account.User))
	h := strings.ToLower(strings.TrimSpace(account.Host))
	if u == "" {
		return false
	}
	for _, sp := range specs {
		if strings.ToLower(sp.user) != u {
			continue
		}
		if sp.host == "" {
			return true
		}
		if strings.ToLower(sp.host) == h {
			return true
		}
	}
	return false
}

func uniqSortedAccounts(in []UserAccount) []UserAccount {
	type key struct {
		u string
		h string
	}
	seen := make(map[key]struct{}, len(in))
	out := make([]UserAccount, 0, len(in))
	for _, a := range in {
		u := strings.TrimSpace(a.User)
		h := strings.TrimSpace(a.Host)
		if u == "" || h == "" {
			continue
		}
		k := key{u: u, h: h}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, UserAccount{User: u, Host: h})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].User == out[j].User {
			return out[i].Host < out[j].Host
		}
		return out[i].User < out[j].User
	})
	return out
}

func ensureSemicolon(s string) string {
	t := strings.TrimSpace(s)
	if t == "" {
		return ""
	}
	if strings.HasSuffix(t, ";") {
		return t
	}
	return t + ";"
}

func databasesNeedlePatterns(db string) []string {
	// Best-effort pattern list untuk deteksi grant yang relevan.
	// Cakup bentuk:
	// - ON `db`.*
	// - ON `db`.`
	// - ON db.*
	// - ON db.`
	db = strings.TrimSpace(db)
	if db == "" {
		return nil
	}
	dbUpper := strings.ToUpper(db)
	return []string{
		" ON `" + dbUpper + "`.", // includes table-level: ON `DB`.`TBL`
		" ON `" + dbUpper + "`.*",
		" ON " + dbUpper + ".*",
		" ON " + dbUpper + ".`",
	}
}
