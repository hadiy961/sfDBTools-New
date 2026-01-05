// File : internal/services/config/appconfig_yaml.go
// Deskripsi : Helper YAML marshal/unmarshal untuk config
// Author : Hadiyatna Muflihun
// Tanggal : 2 Januari 2026
// Last Modified : 5 Januari 2026
package appconfig

import "gopkg.in/yaml.v3"

func MarshalYAML(cfg *Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
