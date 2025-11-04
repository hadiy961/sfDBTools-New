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
