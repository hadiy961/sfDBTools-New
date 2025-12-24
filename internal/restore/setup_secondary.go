package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/restore/display"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

func secondaryDBName(primaryDB string, instance string) string {
	inst := strings.TrimSpace(instance)
	return primaryDB + "_secondary_" + inst
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it == "" {
			continue
		}
		if _, ok := seen[it]; ok {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
	}
	return out
}

func extractSecondaryInstances(databases []string, secondaryPrefix string) []string {
	instances := make([]string, 0)
	for _, db := range databases {
		if !strings.HasPrefix(db, secondaryPrefix) {
			continue
		}
		inst := strings.TrimPrefix(db, secondaryPrefix)
		if inst == "" {
			continue
		}
		instances = append(instances, inst)
	}
	return uniqueStrings(instances)
}

func (s *Service) resolveSecondaryFrom(opts *types.RestoreSecondaryOptions, allowInteractive bool) error {
	from := strings.ToLower(strings.TrimSpace(opts.From))
	if from == "" {
		if allowInteractive {
			selected, err := input.SelectSingleFromList([]string{"file", "primary"}, "Pilih mode restore secondary (source)")
			if err != nil {
				return fmt.Errorf("gagal memilih mode restore secondary: %w", err)
			}
			from = selected
		} else {
			// Backward compatible default for non-interactive runs
			from = "file"
		}
	}
	if from != "file" && from != "primary" {
		if !allowInteractive {
			return fmt.Errorf("nilai --from tidak valid: %s (gunakan: primary atau file)", from)
		}
		selected, err := input.SelectSingleFromList([]string{"file", "primary"}, "Pilih sumber restore secondary")
		if err != nil {
			return fmt.Errorf("gagal memilih --from: %w", err)
		}
		from = selected
	}
	opts.From = from
	return nil
}

func (s *Service) resolveSecondaryClientCode(opts *types.RestoreSecondaryOptions, allowInteractive bool) error {
	if strings.TrimSpace(opts.ClientCode) != "" {
		return nil
	}
	if !allowInteractive {
		return fmt.Errorf("client code wajib diisi (--client-code) pada mode non-interaktif (--force)")
	}
	cc, err := input.AskString("Masukkan client code: ", "", func(ans interface{}) error {
		v, ok := ans.(string)
		if !ok {
			return fmt.Errorf("input tidak valid")
		}
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("client code tidak boleh kosong")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("gagal mendapatkan client code: %w", err)
	}
	opts.ClientCode = strings.TrimSpace(cc)
	return nil
}

func (s *Service) resolveSecondaryEncryptionKey(opts *types.RestoreSecondaryOptions, allowInteractive bool) error {
	// From=file: reuse normal resolver (based on file extension)
	if opts.From == "file" {
		return s.resolveEncryptionKey(opts.File, &opts.EncryptionKey, allowInteractive)
	}

	// From=primary: encryption key may be needed to encrypt the generated backup (if enabled)
	if !s.Config.Backup.Encryption.Enabled {
		return nil
	}
	if strings.TrimSpace(opts.EncryptionKey) != "" {
		return nil
	}
	if !allowInteractive {
		return fmt.Errorf("encryption key wajib diisi (--encryption-key) karena backup encryption aktif")
	}

	k, err := input.AskString("Masukkan encryption key untuk file backup (SFDB_BACKUP_ENCRYPTION_KEY): ", "", func(ans interface{}) error {
		v, ok := ans.(string)
		if !ok {
			return fmt.Errorf("input tidak valid")
		}
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("encryption key tidak boleh kosong")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
	}
	opts.EncryptionKey = strings.TrimSpace(k)
	return nil
}

func (s *Service) resolveSecondaryCompanionFile(allowInteractive bool) error {
	opts := s.RestoreSecondaryOpts
	if opts == nil {
		return fmt.Errorf("restore secondary options tidak tersedia")
	}
	if !opts.IncludeDmart {
		return nil
	}
	if opts.From != "file" {
		// From=primary: companion akan ditangani dari DB, bukan file.
		return nil
	}

	// If companion file explicitly set, validate existence.
	if strings.TrimSpace(opts.CompanionFile) != "" {
		fi, err := os.Stat(opts.CompanionFile)
		if err == nil && !fi.IsDir() {
			return nil
		}
		if !allowInteractive {
			if opts.StopOnError {
				return fmt.Errorf("dmart file (_dmart) tidak ditemukan/invalid: %s", opts.CompanionFile)
			}
			ui.PrintWarning("⚠️  Skip restore companion database (_dmart) karena dmart file invalid")
			opts.IncludeDmart = false
			opts.CompanionFile = ""
			return nil
		}
		// interactive fallback
		ui.PrintWarning(fmt.Sprintf("⚠️  Dmart file tidak valid: %s", opts.CompanionFile))
		opts.CompanionFile = ""
	}

	// Non-interactive + auto-detect disabled => cannot proceed
	if opts.Force && !opts.AutoDetectDmart {
		if opts.StopOnError {
			return fmt.Errorf("auto-detect dmart dimatikan (dmart-detect=false) dan mode non-interaktif (--force) aktif; tentukan dmart via --dmart-file atau set --dmart-include=false")
		}
		ui.PrintWarning("⚠️  Skip restore companion database (_dmart) (non-interaktif, companion tidak ditentukan)")
		opts.IncludeDmart = false
		return nil
	}

	// If auto-detect disabled, select interactively (or skip)
	if !opts.AutoDetectDmart {
		if !allowInteractive {
			return nil
		}
		confirmed, err := input.AskYesNo("Pilih file companion (_dmart) secara manual?", true)
		if err != nil || !confirmed {
			opts.IncludeDmart = false
			return nil
		}
		validExtensions := helper.ValidBackupFileExtensionsForSelection()
		selectedFile, err := input.SelectFileInteractive(filepath.Dir(opts.File), "Masukkan path directory atau file dmart", validExtensions)
		if err != nil {
			return fmt.Errorf("gagal memilih dmart file: %w", err)
		}
		opts.CompanionFile = selectedFile
		return nil
	}

	primaryFile := opts.File
	dir := filepath.Dir(primaryFile)

	// Reuse the same detection strategies as primary restore.
	companionPath, err := s.detectCompanionFromMetadata(primaryFile)
	if err == nil && companionPath != "" {
		opts.CompanionFile = companionPath
		return nil
	}
	companionPath, err = s.detectCompanionByPattern(primaryFile, dir)
	if err == nil && companionPath != "" {
		opts.CompanionFile = companionPath
		return nil
	}
	companionPath, err = s.detectCompanionBySiblingFilename(primaryFile, dir)
	if err == nil && companionPath != "" {
		opts.CompanionFile = companionPath
		return nil
	}

	// Not found
	if !allowInteractive {
		if opts.StopOnError {
			return fmt.Errorf("dmart file (_dmart) tidak ditemukan secara otomatis; gunakan --dmart-file atau set --dmart-include=false")
		}
		ui.PrintWarning("⚠️  Skip restore companion database (_dmart) (companion tidak ditemukan)")
		opts.IncludeDmart = false
		return nil
	}

	ui.PrintWarning("⚠️  Companion (_dmart) file tidak ditemukan secara otomatis")
	selected, err := input.SelectSingleFromList([]string{
		"📁 [ Browse / pilih dmart file ]",
		"⏭️  [ Skip restore companion database (_dmart) ]",
	}, "Pilih tindakan untuk companion (_dmart)")
	if err != nil {
		return fmt.Errorf("gagal memilih opsi companion: %w", err)
	}
	if strings.HasPrefix(selected, "⏭️") {
		opts.IncludeDmart = false
		return nil
	}
	validExtensions := helper.ValidBackupFileExtensionsForSelection()
	selectedFile, err := input.SelectFileInteractive(dir, "Masukkan path directory atau file dmart", validExtensions)
	if err != nil {
		return fmt.Errorf("gagal memilih dmart file: %w", err)
	}
	opts.CompanionFile = selectedFile
	return nil
}

func (s *Service) resolveSecondaryPrimaryDB(ctx context.Context, opts *types.RestoreSecondaryOptions) error {
	cc := strings.TrimSpace(opts.ClientCode)
	if cc == "" {
		return fmt.Errorf("client code kosong")
	}

	// Candidate: allow client-code already prefixed
	ccLower := strings.ToLower(cc)
	if strings.HasPrefix(ccLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(ccLower, consts.PrimaryPrefixBiznet) {
		opts.PrimaryDB = cc
		return nil
	}

	nbc := buildPrimaryTargetDBFromClientCode(consts.PrimaryPrefixNBC, cc)
	biz := buildPrimaryTargetDBFromClientCode(consts.PrimaryPrefixBiznet, cc)

	nbcExists, nbcErr := s.TargetClient.CheckDatabaseExists(ctx, nbc)
	if nbcErr != nil {
		return fmt.Errorf("gagal mengecek database primary (NBC): %w", nbcErr)
	}
	bizExists, bizErr := s.TargetClient.CheckDatabaseExists(ctx, biz)
	if bizErr != nil {
		return fmt.Errorf("gagal mengecek database primary (Biznet): %w", bizErr)
	}

	if nbcExists {
		opts.PrimaryDB = nbc
		if bizExists {
			s.Log.Warnf("Ditemukan 2 primary (%s dan %s); menggunakan %s", nbc, biz, nbc)
		}
		return nil
	}
	if bizExists {
		opts.PrimaryDB = biz
		return nil
	}

	return fmt.Errorf("database primary tidak ditemukan untuk client-code %q (coba: %s atau %s)", cc, nbc, biz)
}

func (s *Service) resolveSecondaryPrefixForFileMode(ctx context.Context, opts *types.RestoreSecondaryOptions) (string, error) {
	cc := strings.TrimSpace(opts.ClientCode)
	if cc == "" {
		return consts.PrimaryPrefixNBC, nil
	}

	// Prefer existing secondary databases on target (if any)
	dbs, err := s.TargetClient.GetNonSystemDatabases(ctx)
	if err == nil {
		needleNBC := consts.PrimaryPrefixNBC + cc + "_secondary_"
		needleBiz := consts.PrimaryPrefixBiznet + cc + "_secondary_"
		for _, db := range dbs {
			if strings.HasPrefix(db, needleNBC) {
				return consts.PrimaryPrefixNBC, nil
			}
		}
		for _, db := range dbs {
			if strings.HasPrefix(db, needleBiz) {
				return consts.PrimaryPrefixBiznet, nil
			}
		}
	}

	// Next: infer from file name
	if strings.TrimSpace(opts.File) != "" {
		return inferPrimaryPrefixFromTargetOrFile("", opts.File), nil
	}

	return consts.PrimaryPrefixNBC, nil
}

func (s *Service) resolveSecondaryInstance(ctx context.Context, opts *types.RestoreSecondaryOptions, primaryDB string, allowInteractive bool) error {
	inst := strings.TrimSpace(opts.Instance)
	if inst != "" {
		target := secondaryDBName(primaryDB, inst)
		if !helper.IsValidDatabaseName(target) {
			return fmt.Errorf("instance menghasilkan nama database tidak valid: %s", target)
		}
		opts.Instance = inst
		return nil
	}

	if !allowInteractive {
		return fmt.Errorf("instance wajib diisi (--instance) pada mode non-interaktif (--force)")
	}

	// Build suggestions from existing databases
	databases, err := s.TargetClient.GetNonSystemDatabases(ctx)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan list database: %w", err)
	}

	secondaryPrefix := primaryDB + "_secondary_"
	instances := extractSecondaryInstances(databases, secondaryPrefix)

	options := []string{"⌨️  [ Input instance baru secara manual ]"}
	if len(instances) > 0 {
		options = append(options, "────────────────────────")
		options = append(options, instances...)
	}

	selected, err := input.SelectSingleFromList(options, "Pilih instance secondary")
	if err != nil {
		return fmt.Errorf("gagal memilih instance: %w", err)
	}

	if selected == "⌨️  [ Input instance baru secara manual ]" {
		val, err := input.AskString("Masukkan instance secondary: ", "training", func(ans interface{}) error {
			v, ok := ans.(string)
			if !ok {
				return fmt.Errorf("input tidak valid")
			}
			v = strings.TrimSpace(v)
			if v == "" {
				return fmt.Errorf("instance tidak boleh kosong")
			}
			target := secondaryDBName(primaryDB, v)
			if !helper.IsValidDatabaseName(target) {
				return fmt.Errorf("instance menghasilkan nama database tidak valid: %s", target)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("gagal mendapatkan instance: %w", err)
		}
		opts.Instance = strings.TrimSpace(val)
		return nil
	}

	if selected == "────────────────────────" {
		return fmt.Errorf("pilihan tidak valid")
	}

	opts.Instance = strings.TrimSpace(selected)
	return nil
}

// SetupRestoreSecondarySession melakukan setup untuk restore secondary session
func (s *Service) SetupRestoreSecondarySession(ctx context.Context) error {
	ui.Headers("Restore Secondary Database")
	allowInteractive := !s.RestoreSecondaryOpts.Force

	// 0. Resolve from
	if err := s.resolveSecondaryFrom(s.RestoreSecondaryOpts, allowInteractive); err != nil {
		return err
	}

	// 1. Resolve file (only when From=file)
	if s.RestoreSecondaryOpts.From == "file" {
		if err := s.resolveBackupFile(&s.RestoreSecondaryOpts.File, allowInteractive); err != nil {
			return fmt.Errorf("gagal resolve file backup: %w", err)
		}
	}

	// 2. Resolve target profile
	if err := s.resolveTargetProfile(&s.RestoreSecondaryOpts.Profile, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve target profile: %w", err)
	}

	// 3. Connect to target database
	if err := s.connectToTargetDatabase(ctx); err != nil {
		return fmt.Errorf("gagal koneksi ke database target: %w", err)
	}

	// 4. Resolve client code
	if err := s.resolveSecondaryClientCode(s.RestoreSecondaryOpts, allowInteractive); err != nil {
		return err
	}

	// 5. Resolve encryption key
	if err := s.resolveSecondaryEncryptionKey(s.RestoreSecondaryOpts, allowInteractive); err != nil {
		return err
	}

	// 5b. Resolve companion file (_dmart) if enabled (file mode)
	if err := s.resolveSecondaryCompanionFile(allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve companion (dmart) file: %w", err)
	}
	// If we ended up with a companion file, ensure encryption key is resolved for that file as well.
	if s.RestoreSecondaryOpts.IncludeDmart && strings.TrimSpace(s.RestoreSecondaryOpts.CompanionFile) != "" {
		if err := s.resolveEncryptionKey(s.RestoreSecondaryOpts.CompanionFile, &s.RestoreSecondaryOpts.EncryptionKey, allowInteractive); err != nil {
			return fmt.Errorf("gagal resolve encryption key untuk dmart file: %w", err)
		}
	}

	// 6. Resolve primary DB (from primary) or determine prefix (from file)
	primaryDB := ""
	if s.RestoreSecondaryOpts.From == "primary" {
		if err := s.resolveSecondaryPrimaryDB(ctx, s.RestoreSecondaryOpts); err != nil {
			return err
		}
		primaryDB = s.RestoreSecondaryOpts.PrimaryDB
	} else {
		prefix, err := s.resolveSecondaryPrefixForFileMode(ctx, s.RestoreSecondaryOpts)
		if err != nil {
			return err
		}
		primaryDB = buildPrimaryTargetDBFromClientCode(prefix, s.RestoreSecondaryOpts.ClientCode)
		s.RestoreSecondaryOpts.PrimaryDB = primaryDB // informational
	}

	// 7. Resolve instance (interactive if empty)
	if err := s.resolveSecondaryInstance(ctx, s.RestoreSecondaryOpts, primaryDB, allowInteractive); err != nil {
		return err
	}

	// 8. Build target secondary db name
	s.RestoreSecondaryOpts.TargetDB = secondaryDBName(primaryDB, s.RestoreSecondaryOpts.Instance)

	// 9. Resolve ticket
	if err := s.resolveTicketNumber(&s.RestoreSecondaryOpts.Ticket, allowInteractive); err != nil {
		return fmt.Errorf("gagal resolve ticket number: %w", err)
	}

	// 10. Interaktif: pilih backup pre-restore & drop target
	if err := s.resolveInteractiveSafetyOptions(&s.RestoreSecondaryOpts.DropTarget, &s.RestoreSecondaryOpts.SkipBackup, allowInteractive); err != nil {
		return err
	}

	// 11. Setup backup options
	if s.RestoreSecondaryOpts.BackupOptions == nil {
		s.RestoreSecondaryOpts.BackupOptions = &types.RestoreBackupOptions{}
	}
	// Backup options diperlukan jika:
	// - user tidak skip backup target, atau
	// - From=primary (karena selalu membuat backup primary)
	if !s.RestoreSecondaryOpts.SkipBackup || s.RestoreSecondaryOpts.From == "primary" {
		s.setupBackupOptions(s.RestoreSecondaryOpts.BackupOptions, s.RestoreSecondaryOpts.EncryptionKey, allowInteractive)
	}

	// 12. Confirmation
	confirmOpts := map[string]string{
		"From":             s.RestoreSecondaryOpts.From,
		"Client Code":      s.RestoreSecondaryOpts.ClientCode,
		"Instance":         s.RestoreSecondaryOpts.Instance,
		"Primary Database": primaryDB,
		"Target Database":  s.RestoreSecondaryOpts.TargetDB,
		"Target Host":      fmt.Sprintf("%s:%d", s.Profile.DBInfo.Host, s.Profile.DBInfo.Port),
		"Drop Target":      fmt.Sprintf("%v", s.RestoreSecondaryOpts.DropTarget),
		"Skip Backup":      fmt.Sprintf("%v", s.RestoreSecondaryOpts.SkipBackup),
		"Ticket Number":    s.RestoreSecondaryOpts.Ticket,
	}
	if s.RestoreSecondaryOpts.From == "file" {
		confirmOpts["Source File"] = s.RestoreSecondaryOpts.File
	}
	if s.RestoreSecondaryOpts.IncludeDmart {
		companionStatus := "Auto-detect"
		if strings.TrimSpace(s.RestoreSecondaryOpts.CompanionFile) != "" {
			companionStatus = filepath.Base(s.RestoreSecondaryOpts.CompanionFile)
		}
		confirmOpts["Companion (dmart)"] = companionStatus
	}
	if s.RestoreSecondaryOpts.BackupOptions != nil && (s.RestoreSecondaryOpts.From == "primary" || !s.RestoreSecondaryOpts.SkipBackup) {
		confirmOpts["Backup Directory"] = s.RestoreSecondaryOpts.BackupOptions.OutputDir
	}

	if !s.RestoreSecondaryOpts.Force {
		if err := display.DisplayConfirmation(confirmOpts); err != nil {
			return err
		}
	}

	// Ensure file path is normalized for display purposes
	if s.RestoreSecondaryOpts.File != "" {
		s.RestoreSecondaryOpts.File = filepath.Clean(s.RestoreSecondaryOpts.File)
	}

	return nil
}
