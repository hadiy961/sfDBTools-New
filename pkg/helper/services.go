package helper

import (
	"sfdbtools/pkg/compress"
	"sfdbtools/pkg/helper/cli"
	cryptokey "sfdbtools/pkg/helper/crypto"
	"sfdbtools/pkg/helper/env"
	"sfdbtools/pkg/helper/file"
	"sfdbtools/pkg/helper/list"
	"sfdbtools/pkg/helper/path"
	"sfdbtools/pkg/helper/profileutil"
	"sfdbtools/pkg/helper/timing"

	"github.com/spf13/cobra"
)

// -------------------- env --------------------

func GetEnvOrDefault(key, defaultValue string) string {
	return env.GetEnvOrDefault(key, defaultValue)
}

func GetEnvOrDefaultInt(key string, defaultValue int) int {
	return env.GetEnvOrDefaultInt(key, defaultValue)
}

func ExpandPath(pathStr string) string {
	return env.ExpandPath(pathStr)
}

// -------------------- crypto --------------------

func ResolveEncryptionKey(existing string, envName string) (string, string, error) {
	return cryptokey.ResolveEncryptionKey(existing, envName)
}

// -------------------- cli flags --------------------

func GetStringFlagOrEnv(cmd *cobra.Command, flagName, envName string) string {
	return cli.GetStringFlagOrEnv(cmd, flagName, envName)
}

func GetSecretStringFlagOrEnv(cmd *cobra.Command, flagName, envName string) (string, error) {
	return cli.GetSecretStringFlagOrEnv(cmd, flagName, envName)
}

func GetIntFlagOrEnv(cmd *cobra.Command, flagName, envName string) int {
	return cli.GetIntFlagOrEnv(cmd, flagName, envName)
}

func GetBoolFlagOrEnv(cmd *cobra.Command, flagName, envName string) bool {
	return cli.GetBoolFlagOrEnv(cmd, flagName, envName)
}

func GetStringSliceFlagOrEnv(cmd *cobra.Command, flagName, envName string) []string {
	return cli.GetStringSliceFlagOrEnv(cmd, flagName, envName)
}

func GetStringArrayFlagOrEnv(cmd *cobra.Command, flagName, envName string) []string {
	return cli.GetStringArrayFlagOrEnv(cmd, flagName, envName)
}

// -------------------- list --------------------

func ListTrimNonEmpty(items []string) []string {
	return list.ListTrimNonEmpty(items)
}

func StringSliceContainsFold(items []string, item string) bool {
	return list.StringSliceContainsFold(items, item)
}

func CSVToCleanList(csv string) []string {
	return list.CSVToCleanList(csv)
}

func ListUnique(items []string) []string {
	return list.ListUnique(items)
}

func ListSubtract(a, b []string) []string {
	return list.ListSubtract(a, b)
}

// -------------------- path --------------------

type PathPatternReplacer = path.PathPatternReplacer

func NewPathPatternReplacer(database string, hostname string, compressionType compress.CompressionType, encrypted bool, isFilename bool) (*PathPatternReplacer, error) {
	return path.NewPathPatternReplacer(database, hostname, compressionType, encrypted, isFilename)
}

func GenerateBackupFilename(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, excludeData bool) (string, error) {
	return path.GenerateBackupFilename(database, mode, hostname, compressionType, encrypted, excludeData)
}

func GenerateBackupFilenameWithCount(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, dbCount int, excludeData bool) (string, error) {
	return path.GenerateBackupFilenameWithCount(database, mode, hostname, compressionType, encrypted, dbCount, excludeData)
}

func GenerateBackupDirectory(baseDir string, structurePattern string, hostname string) (string, error) {
	return path.GenerateBackupDirectory(baseDir, structurePattern, hostname)
}

func GenerateFullBackupPath(baseDir string, structurePattern string, filenamePattern string, database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, excludeData bool) (string, error) {
	return path.GenerateFullBackupPath(baseDir, structurePattern, filenamePattern, database, mode, hostname, compressionType, encrypted, excludeData)
}

// -------------------- file --------------------

func IsEncryptedFile(pathStr string) bool {
	return file.IsEncryptedFile(pathStr)
}

func ValidBackupFileExtensionsForSelection() []string {
	return file.ValidBackupFileExtensionsForSelection()
}

func ExtractDatabaseNameFromFile(filePath string) string {
	return file.ExtractDatabaseNameFromFile(filePath)
}

func IsValidDatabaseName(name string) bool {
	return file.IsValidDatabaseName(name)
}

func ListBackupFilesInDirectory(dir string) ([]string, error) {
	return file.ListBackupFilesInDirectory(dir)
}

func GenerateGrantsFilename(backupFilename string) string {
	return file.GenerateGrantsFilename(backupFilename)
}

func AutoDetectGrantsFile(backupFile string) string {
	return file.AutoDetectGrantsFile(backupFile)
}

func IsBackupFile(filename string) bool {
	return file.IsBackupFile(filename)
}

func StripAllBackupExtensions(filename string) string {
	return file.StripAllBackupExtensions(filename)
}

func ExtractFileExtensions(filename string) (string, []string) {
	return file.ExtractFileExtensions(filename)
}

// -------------------- timing --------------------

type Timer = timing.Timer

func NewTimer() *Timer {
	return timing.NewTimer()
}

// -------------------- profile --------------------

func TrimProfileSuffix(name string) string {
	return profileutil.TrimProfileSuffix(name)
}

func ResolveConfigPath(spec string) (string, string, error) {
	return profileutil.ResolveConfigPath(spec)
}
