package consts

// Database naming conventions.
const (
	PrimaryPrefixNBC    = "dbsf_nbc_"
	PrimaryPrefixBiznet = "dbsf_biznet_"

	SuffixDmart     = "_dmart"
	SuffixTemp      = "_temp"
	SuffixArchive   = "_archive"
	SecondarySuffix = "_secondary"

	// DefaultInitialDatabase is the initial DB to connect to (system DB).
	DefaultInitialDatabase = "mysql"
)
