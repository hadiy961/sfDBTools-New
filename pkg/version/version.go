package version

import "fmt"

// Version, Commit, dan BuildDate bisa di-inject via -ldflags saat build.
// Default value dipakai saat development.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func String() string {
	// Format ringkas dan stabil untuk parsing.
	return fmt.Sprintf("%s (commit=%s, built=%s)", Version, Commit, BuildDate)
}
