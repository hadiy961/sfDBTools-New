package sync

import "time"

// SyncSetting represents a setting to be synchronized.
type SyncSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Category  string    `json:"category"`
	IsLocked  bool      `json:"is_locked"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SyncProfile represents a database profile to be synchronized.
type SyncProfile struct {
	Name          string    `json:"name"`
	EncryptedData string    `json:"encrypted_data"`
	AccountCode   string    `json:"account_code"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SyncJob represents a backup job to be synchronized.
type SyncJob struct {
	Name          string    `json:"name"`
	Enabled       bool      `json:"enabled"`
	Schedule      string    `json:"schedule"`
	Mode          string    `json:"mode"`
	OutputMode    string    `json:"output_mode"`
	IncludeFile   string    `json:"include_file"`
	ProfileName   string    `json:"profile_name"`
	Ticket        string    `json:"ticket"`
	OutputDir     string    `json:"output_dir"`
	RetentionDays int       `json:"retention_days"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SyncPayload represents the full encrypted payload sent to/from the Hub.
type SyncPayload struct {
	ClientCode string        `json:"client_code"`
	Settings   []SyncSetting `json:"settings,omitempty"`
	Profiles   []SyncProfile `json:"profiles,omitempty"`
	Jobs       []SyncJob     `json:"jobs,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
}

// Heartbeat represents the client health data.
type Heartbeat struct {
	ClientCode   string            `json:"client_code"`
	ClientAlias  string            `json:"client_alias"`
	OS           string            `json:"os"`
	Version      string            `json:"version"`
	CPUModel     string            `json:"cpu_model"`
	MemTotal     uint64            `json:"mem_total"`
	MemUsed      uint64            `json:"mem_used"`
	DiskTotal    uint64            `json:"disk_total"`
	DiskFree     uint64            `json:"disk_free"`
	ToolVersions map[string]string `json:"tool_versions"`
	Timestamp    time.Time         `json:"timestamp"`
}
