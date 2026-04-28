package execution

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestTransformSQLStream_Complete(t *testing.T) {
	sourceDB := "db_prod"
	targetDB := "db_dev"

	tests := []struct {
		name   string
		input  string
		verify func(t *testing.T, result string)
	}{
		{
			name:  "db name replacement with backticks",
			input: "/*!50003 TRIGGER `db_prod`.`trg1` AFTER INSERT ON `db_prod`.`tbl` FOR EACH ROW ... */",
			verify: func(t *testing.T, res string) {
				if strings.Contains(res, "`db_prod`.") {
					t.Error("Failed to replace source DB name with backticks")
				}
				if !strings.Contains(res, "`db_dev`.") {
					t.Error("Missing target DB name with backticks")
				}
			},
		},
		{
			name:  "db name replacement with spaces",
			input: "INSERT INTO db_prod.users SELECT * FROM db_prod.old_users",
			verify: func(t *testing.T, res string) {
				if strings.Contains(res, " db_prod.") {
					t.Error("Failed to replace source DB name with space prefix")
				}
				if !strings.Contains(res, " db_dev.") {
					t.Error("Missing target DB name with space prefix")
				}
			},
		},
		{
			name:  "definer stripping style 1 (comment)",
			input: "/*!50003 CREATE*/ /*!50017 DEFINER=`admin`@`%`*/ /*!50003 TRIGGER ... */",
			verify: func(t *testing.T, res string) {
				if strings.Contains(res, "DEFINER") {
					t.Error("Failed to strip DEFINER comment style")
				}
			},
		},
		{
			name:  "definer stripping style 2 (direct)",
			input: "CREATE DEFINER=`papp`@`127.0.0.1` TRIGGER `trg_ls_backup_list_insert` ...",
			verify: func(t *testing.T, res string) {
				if strings.Contains(res, "DEFINER") {
					t.Error("Failed to strip DEFINER direct style")
				}
			},
		},
		{
			name:  "mixed definers and db names",
			input: "DEFINER=`root`@`localhost` VIEW `db_prod`.`vw_test` AS SELECT * FROM db_prod.tbl",
			verify: func(t *testing.T, res string) {
				if strings.Contains(res, "DEFINER") || strings.Contains(res, "db_prod") {
					t.Errorf("Failed to clean mixed line: %s", res)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notify := make(chan string, 10)
			reader := strings.NewReader(tt.input)
			transformed := transformSQLStream(reader, sourceDB, targetDB, notify)

			var out bytes.Buffer
			io.Copy(&out, transformed)
			tt.verify(t, out.String())
		})
	}
}

func TestTransformSQLStream_TableDetection(t *testing.T) {
	input := "\n-- Table structure for table `" + "t_users" + "`\n" +
		"...\n-- Dumping data for table `" + "t_logs" + "`\n" +
		"...\n"
	notify := make(chan string, 10)
	reader := strings.NewReader(input)
	transformed := transformSQLStream(reader, "src", "tgt", notify)

	// Read to trigger goroutine
	io.Copy(io.Discard, transformed)
	close(notify)

	found := make(map[string]bool)
	for tbl := range notify {
		found[tbl] = true
	}

	if !found["t_users"] {
		t.Error("Failed to detect structure table 't_users'")
	}
	if !found["t_logs"] {
		t.Error("Failed to detect data table 't_logs'")
	}
}

func TestTransformSQLStream_LargeBuffer(t *testing.T) {
	// Test if it can handle lines up to 1MB (well within the 10MB limit)
	lineSize := 1 * 1024 * 1024
	largeLine := "INSERT INTO `tbl` VALUES ('" + strings.Repeat("x", lineSize) + "');\n"

	reader := strings.NewReader(largeLine)
	transformed := transformSQLStream(reader, "src", "tgt", nil)

	var out bytes.Buffer
	n, err := io.Copy(&out, transformed)
	if err != nil {
		t.Fatalf("Failed to copy large stream: %v", err)
	}

	if int(n) < lineSize {
		t.Errorf("Truncated output: read %d bytes, want at least %d", n, lineSize)
	}
}
