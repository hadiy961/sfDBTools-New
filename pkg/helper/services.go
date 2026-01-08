package helper

import (
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	backuppath "sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/cli/resolver"
	cryptokey "sfdbtools/internal/services/crypto/helpers"
	"sfdbtools/internal/shared/envx"
	"sfdbtools/internal/shared/listx"
	"sfdbtools/internal/shared/timex"
	"sfdbtools/pkg/compress"
	"sfdbtools/pkg/helper/profileutil"

	"github.com/spf13/cobra"
)

// -------------------- env --------------------

func GetEnvOrDefault(key, defaultValue string) string {
	return envx.GetEnvOrDefault(key, defaultValue)
}

func GetEnvOrDefaultInt(key string, defaultValue int) int {
	return envx.GetEnvOrDefaultInt(key, defaultValue)
}

func ExpandPath(pathStr string) string {
	return envx.ExpandPath(pathStr)
}

// -------------------- crypto --------------------

func ResolveEncryptionKey(existing string, envName string) (string, string, error) {
	return cryptokey.ResolveEncryptionKey(existing, envName)
}

// -------------------- cli flags --------------------

func GetStringFlagOrEnv(cmd *cobra.Command, flagName, envName string) string {
	return resolver.GetStringFlagOrEnv(cmd, flagName, envName)
}

func GetSecretStringFlagOrEnv(cmd *cobra.Command, flagName, envName string) (string, error) {
	return resolver.GetSecretStringFlagOrEnv(cmd, flagName, envName)
}

func GetIntFlagOrEnv(cmd *cobra.Command, flagName, envName string) int {
	return resolver.GetIntFlagOrEnv(cmd, flagName, envName)
}

func GetBoolFlagOrEnv(cmd *cobra.Command, flagName, envName string) bool {
	return resolver.GetBoolFlagOrEnv(cmd, flagName, envName)
}

func GetStringSliceFlagOrEnv(cmd *cobra.Command, flagName, envName string) []string {
	return resolver.GetStringSliceFlagOrEnv(cmd, flagName, envName)
}

func GetStringArrayFlagOrEnv(cmd *cobra.Command, flagName, envName string) []string {
	return resolver.GetStringArrayFlagOrEnv(cmd, flagName, envName)
}

// -------------------- list --------------------

func ListTrimNonEmpty(items []string) []string {
	return listx.ListTrimNonEmpty(items)
}

func StringSliceContainsFold(items []string, item string) bool {
	return listx.StringSliceContainsFold(items, item)
}

func CSVToCleanList(csv string) []string {
	return listx.CSVToCleanList(csv)
}

func ListUnique(items []string) []string {
	return listx.ListUnique(items)
}

func ListSubtract(a, b []string) []string {
	return listx.ListSubtract(a, b)
}

// -------------------- path --------------------

type PathPatternReplacer = backuppath.PathPatternReplacer

func NewPathPatternReplacer(database string, hostname string, compressionType compress.CompressionType, encrypted bool, isFilename bool) (*PathPatternReplacer, error) {
	return backuppath.NewPathPatternReplacer(database, hostname, compressionType, encrypted, isFilename)
}

func GenerateBackupFilename(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, excludeData bool) (string, error) {
	return backuppath.GenerateBackupFilename(database, mode, hostname, compressionType, encrypted, excludeData)
}

func GenerateBackupFilenameWithCount(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, dbCount int, excludeData bool) (string, error) {
	return backuppath.GenerateBackupFilenameWithCount(database, mode, hostname, compressionType, encrypted, dbCount, excludeData)
}

func GenerateBackupDirectory(baseDir string, structurePattern string, hostname string) (string, error) {
	return backuppath.GenerateBackupDirectory(baseDir, structurePattern, hostname)
}

func GenerateFullBackupPath(baseDir string, structurePattern string, filenamePattern string, database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, excludeData bool) (string, error) {
	return backuppath.GenerateFullBackupPath(baseDir, structurePattern, filenamePattern, database, mode, hostname, compressionType, encrypted, excludeData)
}

// -------------------- file --------------------

func IsEncryptedFile(pathStr string) bool {
	return backupfile.IsEncryptedFile(pathStr)
}

func ValidBackupFileExtensionsForSelection() []string {
	return backupfile.ValidBackupFileExtensionsForSelection()
}

func ExtractDatabaseNameFromFile(filePath string) string {
	return backupfile.ExtractDatabaseNameFromFile(filePath)
}

func IsValidDatabaseName(name string) bool {
	return backupfile.IsValidDatabaseName(name)
}

func ListBackupFilesInDirectory(dir string) ([]string, error) {
	return backupfile.ListBackupFilesInDirectory(dir)
}

func GenerateGrantsFilename(backupFilename string) string {
	return backupfile.GenerateGrantsFilename(backupFilename)
}

func AutoDetectGrantsFile(backupFile string) string {
	return backupfile.AutoDetectGrantsFile(backupFile)
}

func IsBackupFile(filename string) bool {
	return backupfile.IsBackupFile(filename)
}

func StripAllBackupExtensions(filename string) string {
	return backupfile.StripAllBackupExtensions(filename)
}

func ExtractFileExtensions(filename string) (string, []string) {
	return backupfile.ExtractFileExtensions(filename)
}

// -------------------- timing --------------------

type Timer = timex.Timer

func NewTimer() *Timer {
	return timex.NewTimer()
}

// -------------------- profile --------------------

func TrimProfileSuffix(name string) string {
	return profileutil.TrimProfileSuffix(name)
}

func ResolveConfigPath(spec string) (string, string, error) {
	return profileutil.ResolveConfigPath(spec)
}
