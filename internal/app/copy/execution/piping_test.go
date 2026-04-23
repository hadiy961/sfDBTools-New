package execution

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestTransformSQLStream(t *testing.T) {
	sourceDB := "db_prod"
	targetDB := "db_dev"
	
	input := `
-- Table structure for table ` + "`" + `users` + "`" + `
CREATE TABLE ` + "`" + `users` + "`" + ` (id int);
/*!50003 CREATE*/ /*!50017 DEFINER=` + "`" + `admin` + "`" + `@` + "`" + `%` + "`" + `*/ /*!50003 TRIGGER ` + "`" + `db_prod` + "`" + `.` + "`" + `trg1` + "`" + ` AFTER INSERT ON db_prod.users FOR EACH ROW BEGIN END */;
-- Dumping data for table ` + "`" + `users` + "`" + `
INSERT INTO ` + "`" + `users` + "`" + ` VALUES (1);
`
	
	notifyChan := make(chan string, 10)
	reader := strings.NewReader(input)
	transformed := transformSQLStream(reader, sourceDB, targetDB, notifyChan)
	
	var output bytes.Buffer
	_, err := io.Copy(&output, transformed)
	if err != nil {
		t.Fatalf("Failed to copy transformed stream: %v", err)
	}
	
	result := output.String()
	
	// 1. Verify DB Name Replacement
	if strings.Contains(result, "`db_prod`.") {
		t.Errorf("Result still contains source DB name with backticks: %s", result)
	}
	if !strings.Contains(result, "`db_dev`.") {
		t.Errorf("Result does not contain target DB name with backticks: %s", result)
	}
	if !strings.Contains(result, " db_dev.") {
		t.Errorf("Result does not contain target DB name with space prefix: %s", result)
	}

	// 2. Verify DEFINER Stripping
	if strings.Contains(result, "DEFINER") {
		t.Errorf("Result still contains DEFINER clause: %s", result)
	}

	// 3. Verify Table Notification
	close(notifyChan)
	var foundTable bool
	for tbl := range notifyChan {
		if tbl == "users" {
			foundTable = true
		}
	}
	if !foundTable {
		t.Errorf("NotifyChan did not receive 'users' table")
	}
}
