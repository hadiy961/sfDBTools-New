package consts

// File : pkg/consts/consts_profile.go
// Deskripsi : Konstanta yang digunakan oleh fitur profile (prompt, header, dan pesan umum)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

// =============================================================================
// Header + Success message
// =============================================================================

const (
	ProfileHeaderCreate = "Create Database Profile"
	ProfileHeaderShow   = "Show Database Profile"
	ProfileHeaderEdit   = "Edit Database Profile"
	ProfileHeaderDelete = "Delete Database Profile"

	ProfileSuccessCreated = "✓ Profile berhasil dibuat"
	ProfileSuccessUpdated = "✓ Profile berhasil diupdate"
	ProfileSuccessDeleted = "✓ Profile berhasil dihapus"
)

// =============================================================================
// Prompt labels & pesan umum (profile wizard)
// =============================================================================

const (
	ProfileNonInteractiveQuiet     = "non-interaktif (--quiet):"
	ProfileMsgNonInteractivePrefix = "mode " + ProfileNonInteractiveQuiet + " "

	// Mode strings
	ProfileModeCreate = "create"
	ProfileModeShow   = "show"
	ProfileModeEdit   = "edit"
	ProfileModeDelete = "delete"

	// UI text / action labels
	ProfileUIHeaderCreate         = "Pembuatan Profil Baru"
	ProfileUIHeaderEdit           = "Database Configuration Editing"
	ProfileUIHeaderShow           = "Display User Profile Details"
	ProfileUIHeaderDelete         = "Delete Database Configurations"
	ProfilePromptAction           = "Aksi:"
	ProfileActionEditData         = "Ubah data"
	ProfileActionCancel           = "Batalkan"
	ProfileMsgChangeSummaryPrefix = "Ringkasan Perubahan : "

	// Wizard headers / labels
	ProfileWizardSubHeaderConfigName        = "Please provide the configuration name:"
	ProfileWizardLabelConfigName            = "Configuration Name"
	ProfileWizardDefaultConfigName          = "localhost_3306"
	ProfileWizardSubHeaderProfileDetails    = "Please provide the following database profile details:"
	ProfileWizardLogStarted                 = "Wizard profil dimulai..."
	ProfileWizardPromptSelectExistingConfig = "Select Existing Configuration File"
	ProfileWizardMsgRestart                 = "Penyimpanan dibatalkan. Memulai ulang wizard..."
	ProfileWizardMsgConfirmAccepted         = "Konfirmasi diterima. Mempersiapkan enkripsi dan penyimpanan..."
	ProfileErrProfileNameEmpty              = "nama konfigurasi tidak boleh kosong"
	ProfileTipKeepCurrentDBPasswordUpdate   = "💡 Tekan Enter untuk mempertahankan password saat ini, atau ketik password baru untuk mengupdate."

	ProfileMsgConfigWillBeSavedAsPrefix = "Konfigurasi akan disimpan sebagai : "
	ProfileMsgNoFieldsSelected          = "Tidak ada field yang dipilih untuk diubah."

	ProfilePromptSelectFieldsToChange = "Pilih data yang ingin diubah:"
	ProfilePromptConfirmSave          = "Apakah Anda ingin menyimpan konfigurasi ini?"
	ProfilePromptConfirmRetry         = "Apakah Anda ingin mengulang proses?"
	ProfilePromptRetryInputConfig     = "Apakah Anda ingin mengulang input konfigurasi?"

	// Shared labels (dipakai ulang oleh prompt/field/display)
	ProfileLabelDBHost          = "Database Host"
	ProfileLabelDBPort          = "Database Port"
	ProfileLabelDBUser          = "Database User"
	ProfileLabelDBPassword      = "Database Password"
	ProfileLabelSSHHost         = "SSH Host"
	ProfileLabelSSHPort         = "SSH Port"
	ProfileLabelSSHUser         = "SSH User"
	ProfileLabelSSHPassword     = "SSH Password"
	ProfileLabelSSHIdentityFile = "SSH Identity File"
	ProfileLabelLocalPort       = "Local Port"
	ProfileLabelSSHLocalPort    = "SSH Local Port"

	ProfileSuffixOptional     = " (opsional)"
	ProfileSuffixEmptyDefault = " (kosong = default)"
	ProfileSuffixZeroAuto     = " (0 = otomatis)"

	ProfilePromptUseSSHTunnel            = "Gunakan SSH tunnel untuk koneksi database?"
	ProfilePromptSSHHost                 = "SSH Host (bastion)"
	ProfilePromptSSHUser                 = ProfileLabelSSHUser + ProfileSuffixEmptyDefault
	ProfilePromptSSHPasswordOptional     = ProfileLabelSSHPassword + ProfileSuffixOptional
	ProfilePromptSSHIdentityFileOptional = ProfileLabelSSHIdentityFile + ProfileSuffixOptional
	ProfilePromptSSHLocalPort            = ProfileLabelLocalPort + ProfileSuffixZeroAuto

	ProfileTipKeepCurrentDBPassword  = "💡 Tekan Enter untuk mempertahankan password saat ini."
	ProfileTipKeepCurrentSSHPassword = "💡 Tekan Enter untuk mempertahankan SSH password saat ini."
	ProfileTipSSHPasswordOptional    = "💡 SSH password opsional. Kosongkan jika menggunakan key/agent."
)

// =============================================================================
// Error / warning messages (profile)
// =============================================================================

const (
	ProfileErrConfigFileNotFoundFmt               = "file konfigurasi tidak ditemukan: %s"
	ProfileWarnConfigFileNotFoundFmt              = "File konfigurasi '%s' tidak ditemukan."
	ProfileWarnLoadFileContentFailedProceedFmt    = "Gagal memuat isi file '%s': %v. Lanjut dengan data minimum."
	ProfileErrLoadSnapshotUnavailable             = "load snapshot tidak tersedia"
	ProfileErrWizardRunnerUnavailable             = "internal error: wizard runner tidak tersedia"
	ProfileErrNonInteractiveProfileRequired       = "flag --profile wajib disertakan pada mode non-interaktif"
	ProfileErrNewNameEmpty                        = "nama baru tidak boleh kosong"
	ProfileErrNewNameContainsPath                 = "nama baru tidak boleh berisi path"
	ProfileWarnEnvVarMissingOrEmptyFmt            = "Environment variable %s tidak ditemukan atau kosong. Silakan atur %s atau ketik password."
	ProfileErrResolveEncryptionKeyUnavailable     = "resolve kunci enkripsi tidak tersedia"
	ProfileErrNonInteractiveProfileFlagRequired   = "mode non-interaktif (--quiet): flag --profile wajib disertakan"
	ProfileErrPromptSelectorUnavailable           = "tidak ada prompt selector"
	ProfileErrNoSnapshotToShow                    = "tidak ada snapshot konfigurasi untuk ditampilkan"
	ProfileWarnDBConnectFailedPrefix              = "Koneksi database gagal: "
	ProfileErrNonInteractiveProfileKeyRequiredFmt = "mode non-interaktif (--quiet): --profile-key wajib diisi atau set ENV %s atau %s: %w"
	ProfileErrDBPasswordRequiredNonInteractiveFmt = "password database wajib diisi via --password atau env %s (mode non-interaktif --quiet): %w"
	ProfileWarnDBPasswordPrompting                = "Password database belum diisi; meminta input password..."
	ProfilePromptDBPasswordForUserFmt             = "Password untuk user (%s) : "
	ProfileErrConfigNameExistsFmt                 = "nama konfigurasi '%s' sudah ada. Silakan pilih nama lain"
	ProfileErrConfigFileNotFoundChooseOtherFmt    = "file konfigurasi '%s' tidak ditemukan. Silakan pilih nama lain"
	ProfileErrOriginalConfigFileNotFoundFmt       = "file konfigurasi asli '%s' tidak ditemukan"
	ProfileErrTargetConfigNameExistsFmt           = "nama konfigurasi tujuan '%s' sudah ada. Silakan pilih nama lain"
	ProfileErrProfileInfoNil                      = "informasi profil tidak boleh kosong"
	ProfileErrProfileNameEmptyAlt                 = "nama profil tidak boleh kosong"
	ProfileErrValidateDBInfoFailedFmt             = "validasi informasi database gagal: %w"
	ProfileErrSSHTunnelHostEmpty                  = "ssh tunnel aktif tapi ssh host kosong"
	ProfileErrDBInfoNil                           = "informasi database tidak boleh kosong"
	ProfileErrDBHostEmpty                         = "host database tidak boleh kosong"
	ProfileErrDBPortInvalidFmt                    = "port database tidak valid: %d"
	ProfileErrDBUserEmpty                         = "user database tidak boleh kosong"
	ProfileErrGetEncryptionPasswordFailedFmt      = "gagal mendapatkan password enkripsi: %w"
	ProfileFmtFileSizeBytes                       = "%d bytes"
)

// =============================================================================
// Retry / cancel messages
// =============================================================================

const (
	ProfileMsgRetryCreate     = "Mengulang proses pembuatan profil..."
	ProfileMsgCreateCancelled = "Proses pembuatan profil dibatalkan."
	ProfileMsgRetryEdit       = "Mengulang proses pengeditan profil..."
	ProfileMsgEditCancelled   = "Proses pengeditan profil dibatalkan."

	ProfileDeleteNoValidProfiles        = "Tidak ada profil valid yang ditemukan untuk dihapus."
	ProfileDeleteForceDeletedFmt        = "Berhasil menghapus (force): %s"
	ProfileDeleteDeletedFmt             = "Berhasil menghapus: %s"
	ProfileDeleteFailedFmt              = "Gagal menghapus: %s (%v)"
	ProfileDeleteWillDeleteHeader       = "Akan menghapus profil berikut:"
	ProfileDeleteConfirmPrefix          = "Anda yakin ingin menghapus %d "
	ProfileDeleteConfirmCountFmt        = ProfileDeleteConfirmPrefix + "profil ini?"
	ProfileDeleteCancelledByUser        = "Penghapusan dibatalkan oleh pengguna."
	ProfileDeleteReadConfigDirFailedFmt = "gagal membaca direktori konfigurasi: %w"
	ProfileDeleteNoConfigFiles          = "Tidak ada file konfigurasi untuk dihapus."
	ProfileDeleteSelectFilesPrompt      = "Pilih file konfigurasi yang akan dihapus:"
	ProfileDeleteNoFilesSelected        = "Tidak ada file terpilih untuk dihapus."
	ProfileDeleteConfirmFilesCountFmt   = ProfileDeleteConfirmPrefix + "file?"
	ProfileDeleteListPrefix             = " - "
)

// =============================================================================
// SaveProfile strings (internal/profile/executor/save.go)
// =============================================================================

const (
	ProfileSaveModeEdit   = "Edit"
	ProfileSaveModeCreate = "Create"

	ProfileErrCreateConfigDirFailedFmt    = "gagal membuat direktori konfigurasi: %w"
	ProfileErrFormatINIUnavailable        = "format config INI tidak tersedia"
	ProfileErrEncryptionKeyUnavailableFmt = "kunci enkripsi tidak tersedia: %w"
	ProfileErrEncryptConfigFailedFmt      = "gagal mengenkripsi konten konfigurasi: %w"
	ProfileErrWriteConfigBase             = "gagal menyimpan file konfigurasi"
	ProfileErrWriteNewConfigFailedFmt     = ProfileErrWriteConfigBase + " baru: %w"
	ProfileErrWriteConfigFailedFmt        = ProfileErrWriteConfigBase + ": %w"

	ProfileSavePromptContinueDespiteDBFail = "\nKoneksi database gagal. Apakah Anda tetap ingin menyimpan konfigurasi ini?"
	ProfileSaveWarnSavingWithInvalidConn   = "⚠️  PERINGATAN: Menyimpan konfigurasi dengan koneksi database yang tidak valid."
	ProfileWarnSavedButDeleteOldFailedFmt  = "Berhasil menyimpan '%s' tetapi gagal menghapus file lama '%s': %v"
	ProfileSuccessSavedRenamedFmt          = "File konfigurasi berhasil disimpan sebagai '%s' (rename dari '%s')."
	ProfileSuccessSavedSafelyFmt           = "File konfigurasi '%s' berhasil disimpan dengan aman."
)

// =============================================================================
// Log strings (internal/profile)
// =============================================================================

const (
	ProfileLogStartProcessFmt             = "Memulai proses profile - %s"
	ProfileLogStartProcessWithPrefixFmt   = "[%s] Memulai proses profile - %s"
	ProfileLogStartOperationWithPrefixFmt = "[%s] Memulai profile operation dengan mode: %s"
	ProfileLogConfigLoadedFromFmt         = "Memuat konfigurasi dari: %s Name: %s"
	ProfileLogEncryptionKeyObtained       = "Password enkripsi berhasil didapatkan."
	ProfileLogLoadConfigDetailsFailedFmt  = "Gagal memuat isi detail konfigurasi: %v"
	ProfileLogStartSave                   = "Memulai proses penyimpanan file konfigurasi..."
	ProfileLogDBConnectionValid           = "Koneksi database valid."
	ProfileLogModeInteractiveEnabled      = "Mode interaktif diaktifkan."
	ProfileLogModeNonInteractiveEnabled   = "Mode non-interaktif diaktifkan."
	ProfileLogModeNonInteractiveShort     = "Mode non-interaktif."
	ProfileLogValidatingParams            = "Memvalidasi parameter..."
	ProfileLogValidationFailedFmt         = "Validasi parameter gagal: %v"
	ProfileLogValidationSuccess           = "Validasi parameter berhasil."
	ProfileLogCreateStarted               = "Memulai proses pembuatan profil baru..."
	ProfileLogCreateSuccess               = "Profil baru berhasil dibuat."
	ProfileLogEditCancelledByUser         = "Proses pengeditan konfigurasi dibatalkan oleh pengguna."
	ProfileLogEditFailedFmt               = "Proses pengeditan konfigurasi gagal: %v"
	ProfileLogConfigFileFoundTryLoad      = "File ditemukan. Mencoba memuat konten..."
	ProfileLogConfigFileLoaded            = "File konfigurasi berhasil dimuat."
	ProfileLogWizardInteractiveFinished   = "Wizard interaktif selesai."
	ProfileLogDeleteFileFailedFmt         = "Gagal menghapus file %s: %v"
	ProfileLogUnknownProfileTypeInService = "Tipe profil tidak dikenali dalam Service"

	ProfileLogPrefixCreate = "profile-create"
	ProfileLogPrefixShow   = "profile-show"
	ProfileLogPrefixEdit   = "profile-edit"
	ProfileLogPrefixDelete = "profile-delete"
)

// Label field untuk multi-select edit (wizard)
const (
	ProfileFieldName            = "Nama profil"
	ProfileFieldSSHTunnelToggle = "SSH Tunnel (enable/disable)"
)

// =============================================================================
// Output messages
// =============================================================================

const (
	ProfileMsgConfigSavedAtPrefix = "File konfigurasi tersimpan di : "
)

// =============================================================================
// CLI help text (cmd/profile)
// =============================================================================

const (
	ProfileCLIAutoInteractiveSuffix           = " (otomatis interaktif kecuali --quiet)"
	ProfileCLIModeNonInteractiveHeader        = "Mode " + ProfileNonInteractiveQuiet
	ProfileCLINonInteractiveEnvProfileKeyNote = "(atau ENV SFDB_TARGET_PROFILE_KEY/SFDB_SOURCE_PROFILE_KEY)."
)

// =============================================================================
// Display strings (internal/profile/display)
// =============================================================================

const (
	ProfileDisplayShowPrefix            = "Menampilkan Profil: "
	ProfileDisplayCreatePrefix          = "Konfigurasi Database Baru: "
	ProfileDisplayNoConfigLoaded        = "Tidak ada konfigurasi yang dimuat untuk ditampilkan."
	ProfileDisplayNoProfileInfo         = "Tidak ada informasi profil untuk ditampilkan."
	ProfileDisplayNoChangeInfo          = "Tidak ada informasi perubahan (tidak ada snapshot asli)."
	ProfileDisplayNoChangesDetected     = "Tidak ada perubahan yang terdeteksi pada konfigurasi."
	ProfileDisplayNoFileForVerify       = "Tidak ada file yang terkait untuk memverifikasi password."
	ProfileDisplayVerifyKeyPrompt       = "Masukkan ulang encryption key untuk verifikasi: "
	ProfileDisplayVerifyKeyFailedPrefix = "Gagal mendapatkan encryption key: "
	ProfileDisplayNoKeyProvided         = "Tidak ada encryption key yang diberikan. Tidak dapat menampilkan password asli."
	ProfileDisplayInvalidKeyOrCorrupt   = "Enkripsi key salah atau file rusak. Tidak dapat menampilkan password asli."
	ProfileDisplayRevealedPasswordTitle = "Revealed Password"

	ProfileDisplayTableHeaderNo     = "No"
	ProfileDisplayTableHeaderField  = "Field"
	ProfileDisplayTableHeaderValue  = "Value"
	ProfileDisplayTableHeaderBefore = "Before"
	ProfileDisplayTableHeaderAfter  = "After"

	ProfileDisplayFieldName         = "Nama"
	ProfileDisplayFieldFilePath     = "File Path"
	ProfileDisplayFieldHost         = "Host"
	ProfileDisplayFieldPort         = "Port"
	ProfileDisplayFieldUser         = "User"
	ProfileDisplayFieldPassword     = "Password"
	ProfileDisplayFieldSSHTunnel    = "SSH Tunnel"
	ProfileDisplayFieldFileSize     = "File Size"
	ProfileDisplayFieldLastModified = "Last Modified"

	ProfileDisplayStateNotSet = "(not set)"
	ProfileDisplayStateSet    = "(set)"
	ProfileDisplaySSHDisabled = "disabled"
	ProfileDisplaySSHEnabled  = "enabled"
)
