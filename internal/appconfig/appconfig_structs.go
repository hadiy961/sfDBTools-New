// File : internal/appconfig/appconfig_types.go
// Deskripsi : Struct untuk konfigurasi aplikasi yang di-load dari file YAML
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package appconfig

// Config adalah struktur level atas yang memegang semua bagian konfigurasi.
// Tag 'yaml' digunakan untuk memetakan field Go ke kunci di file YAML.
type Config struct {
	Backup      BackupConfig      `yaml:"backup"`
	ConfigDir   ConfigDirConfig   `yaml:"config_dir"`
	General     GeneralConfig     `yaml:"general"`
	Log         LogConfig         `yaml:"log"`
	Mariadb     MariadbConfig     `yaml:"mariadb"`
	SystemUsers SystemUsersConfig `yaml:"system_users"`
}

// Struct untuk bagian 'backup'
type BackupConfig struct {
	Compression   CompressionConfig  `yaml:"compression"`
	MysqlDumpArgs string             `yaml:"mysqldump_args"`
	Exclude       ExcludeConfig      `yaml:"exclude"`
	Include       IncludeConfig      `yaml:"include"`
	Cleanup       CleanupConfig      `yaml:"cleanup"`
	Encryption    EncryptionConfig   `yaml:"encryption"`
	Output        OutputConfig       `yaml:"output"`
	Verification  VerificationConfig `yaml:"verification"`
	Replication   ReplicationConfig  `yaml:"replication"`
}

type IncludeConfig struct {
	Databases      []string `yaml:"databases"`
	File           string   `yaml:"file"`
	IncludeDmart   bool     `yaml:"include_dmart"`
	IncludeTemp    bool     `yaml:"include_temp"`
	IncludeArchive bool     `yaml:"include_archive"`
}

type CompressionConfig struct {
	Type    string `yaml:"type"`
	Level   int    `yaml:"level"`
	Enabled bool   `yaml:"enabled"`
}

type ExcludeConfig struct {
	Databases       []string `yaml:"databases"`
	User            bool     `yaml:"user"`
	SystemDatabases bool     `yaml:"system_databases"`
	Data            bool     `yaml:"data"`
	Empty           bool     `yaml:"empty"`
	File            string   `yaml:"file"`
}

type CleanupConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Schedule string `yaml:"schedule"`
	Days     int    `yaml:"days"`
}

type EncryptionConfig struct {
	Enabled bool   `yaml:"enabled"`
	Key     string `yaml:"key"`
}

type OutputConfig struct {
	BaseDirectory string `yaml:"base_directory"`
	CleanupTemp   bool   `yaml:"cleanup_temp"`
	Structure     struct {
		CreateSubdirs bool   `yaml:"create_subdirs"`
		Pattern       string `yaml:"pattern"`
	} `yaml:"structure"`
	SaveBackupInfo bool `yaml:"save_backup_info"`
}

type VerificationConfig struct {
	DiskSpaceCheck bool `yaml:"disk_space_check"`
}

type ReplicationConfig struct {
	CaptureGtid         bool   `yaml:"capture_gtid"`
	ReplicationUser     string `yaml:"replication_user"`
	ReplicationPassword string `yaml:"replication_password"`
}

// Struct untuk bagian 'config_dir'
type ConfigDirConfig struct {
	DatabaseProfile        string `yaml:"database_profile"`
	DatabaseList           string `yaml:"database_list"`
	MariadbConfigTemplates string `yaml:"mariadb_config_templates"`
}

// Struct untuk bagian 'general'
type GeneralConfig struct {
	AppName    string `yaml:"app_name"`
	Author     string `yaml:"author"`
	ClientCode string `yaml:"client_code"`
	Locale     struct {
		DateFormat string `yaml:"date_format"`
		TimeFormat string `yaml:"time_format"`
		Timezone   string `yaml:"timezone"`
	} `yaml:"locale"`
	Version string `yaml:"version"`
}

// Struct untuk bagian 'log'
type LogConfig struct {
	Format string `yaml:"format"`
	Level  string `yaml:"level"`
	Output struct {
		Console struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"console"`
		File struct {
			Dir             string `yaml:"dir"`
			Enabled         bool   `yaml:"enabled"`
			FilenamePattern string `yaml:"filename_pattern"`
			Rotation        struct {
				CompressOld   bool   `yaml:"compress_old"`
				Daily         bool   `yaml:"daily"`
				MaxSize       string `yaml:"max_size"`
				RetentionDays int    `yaml:"retention_days"` // Menggunakan time.Duration untuk konversi
			} `yaml:"rotation"`
		} `yaml:"file"`
	} `yaml:"output"`
	Timezone string `yaml:"timezone"`
}

// Struct untuk bagian 'mariadb'
type MariadbConfig struct {
	BinlogDir           string `yaml:"binlog_dir"`
	ConfigDir           string `yaml:"config_dir"`
	DataDir             string `yaml:"data_dir"`
	EncryptionKeyFile   string `yaml:"encryption_key_file"`
	InnodbEncryptTables bool   `yaml:"innodb_encrypt_tables"`
	LogDir              string `yaml:"log_dir"`
	Port                int    `yaml:"port"`
	ServerID            int    `yaml:"server_id"`
	Version             string `yaml:"version"`
}

// Struct untuk bagian 'system_users'
type SystemUsersConfig struct {
	Users []string `yaml:"users"`
}
