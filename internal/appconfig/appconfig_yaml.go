// File : internal/appconfig/appconfig_yaml.go
// Deskripsi : Helper YAML marshal/unmarshal untuk config
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified : 2026-01-02
package appconfig

import "gopkg.in/yaml.v3"

func MarshalYAML(cfg *Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
