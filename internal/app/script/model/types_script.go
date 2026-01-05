package types

type ScriptEncryptOptions struct {
	FilePath      string
	EncryptionKey string
	Mode          string
	OutputPath    string
	DeleteSource  bool
}

type ScriptRunOptions struct {
	FilePath      string
	EncryptionKey string
	Args          []string
}

type ScriptExtractOptions struct {
	FilePath      string
	EncryptionKey string
	OutDir        string
}

type ScriptInfoOptions struct {
	FilePath      string
	EncryptionKey string
}
