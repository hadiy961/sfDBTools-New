package fsops

import "os"

// ReadDirFiles membaca nama-nama file dalam direktori yang diberikan.
func ReadDirFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

// Cek apakah file atau direktori ada
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CheckDirExists mengecek apakah direktori ada
func CheckDirExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// CreateDirIfNotExist membuat direktori jika belum ada
func CreateDirIfNotExist(dir string) error {
	// Cek apakah direktori sudah ada
	exists, err := CheckDirExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		// Buat direktori beserta parent-nya
		if err := CreateDir(dir); err != nil {
			return err
		}
	}
	return nil
}

// CreateDir membuat direktori beserta parent-nya jika belum ada
func CreateDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}
