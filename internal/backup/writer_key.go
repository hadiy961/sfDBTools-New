package backup

import (
	"fmt"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
)

// resolveEncryptionKeyIfNeeded resolve encryption key jika encryption enabled
func (s *Service) resolveEncryptionKeyIfNeeded() (string, error) {
	if !s.BackupDBOptions.Encryption.Enabled {
		return "", nil
	}

	resolvedKey, source, err := helper.ResolveEncryptionKey(
		s.BackupDBOptions.Encryption.Key,
		consts.ENV_BACKUP_ENCRYPTION_KEY,
	)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
	}

	s.Log.Debugf("Kunci enkripsi didapat dari: %s", source)
	return resolvedKey, nil
}
