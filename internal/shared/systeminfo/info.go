package systeminfo

import (
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

// Info holds system resource information.
type Info struct {
	CPUModel    string
	CPUUsage    float64
	MemoryTotal uint64
	MemoryUsed  uint64
	DiskTotal   uint64
	DiskFree    uint64
}

// GetSystemInfo gathers current CPU, Memory, and Disk usage.
func GetSystemInfo(backupPath string) (*Info, error) {
	info := &Info{}

	// CPU Info (Linux specific)
	if runtime.GOOS == "linux" {
		out, _ := exec.Command("grep", "-m", "1", "model name", "/proc/cpuinfo").Output()
		info.CPUModel = strings.TrimSpace(strings.TrimPrefix(string(out), "model name	: "))
	} else {
		info.CPUModel = runtime.GOARCH
	}

	// Memory Info
	var sysInfo syscall.Sysinfo_t
	if err := syscall.Sysinfo(&sysInfo); err == nil {
		info.MemoryTotal = sysInfo.Totalram * uint64(sysInfo.Unit)
		info.MemoryUsed = (sysInfo.Totalram - sysInfo.Freeram) * uint64(sysInfo.Unit)
	}

	// Disk Info
	var stat syscall.Statfs_t
	if err := syscall.Statfs(backupPath, &stat); err == nil {
		info.DiskTotal = stat.Blocks * uint64(stat.Bsize)
		info.DiskFree = stat.Bfree * uint64(stat.Bsize)
	}

	return info, nil
}

// GetToolVersions returns versions of various database tools.
func GetToolVersions() map[string]string {
	versions := make(map[string]string)
	tools := map[string]string{
		"mariadb":      "mariadb --version",
		"mysql":        "mysql --version",
		"mysqldump":    "mysqldump --version",
		"mariadb-dump": "mariadb-dump --version",
		"mariabackup":  "mariabackup --version",
	}

	for name, cmd := range tools {
		parts := strings.Split(cmd, " ")
		out, err := exec.Command(parts[0], parts[1:]...).Output()
		if err == nil {
			versions[name] = strings.TrimSpace(string(out))
		} else {
			versions[name] = "not installed"
		}
	}

	return versions
}
