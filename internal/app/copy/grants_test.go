package copy

import (
	"testing"
)

func TestTransformGrantsSQL(t *testing.T) {
	s := &Service{}
	
	tests := []struct {
		name     string
		sql      string
		sourceDB string
		targetDB string
		want     string
	}{
		{
			name:     "standard quoted grant",
			sql:      "GRANT ALL PRIVILEGES ON `db_old`.* TO `user`@`localhost`",
			sourceDB: "db_old",
			targetDB: "db_new",
			want:     "GRANT ALL PRIVILEGES ON `db_new`.* TO `user`@`localhost`",
		},
		{
			name:     "standard unquoted grant",
			sql:      "GRANT SELECT ON db_old.* TO `user`@`localhost`",
			sourceDB: "db_old",
			targetDB: "db_new",
			want:     "GRANT SELECT ON db_new.* TO `user`@`localhost`",
		},
		{
			name:     "multiple occurrences",
			sql:      "GRANT ALL ON `db_old`.* TO 'u'@'%'; GRANT SELECT ON `db_old`.`tbl` TO 'u'@'%';",
			sourceDB: "db_old",
			targetDB: "db_new",
			want:     "GRANT ALL ON `db_new`.* TO 'u'@'%'; GRANT SELECT ON `db_new`.`tbl` TO 'u'@'%';",
		},
		{
			name:     "mixed backticks and direct",
			sql:      "GRANT ALL ON `db_old`.* TO 'u'; GRANT ALL ON db_old.* TO 'u';",
			sourceDB: "db_old",
			targetDB: "db_new",
			want:     "GRANT ALL ON `db_new`.* TO 'u'; GRANT ALL ON db_new.* TO 'u';",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.transformGrantsSQL(tt.sql, tt.sourceDB, tt.targetDB)
			if got != tt.want {
				t.Errorf("transformGrantsSQL() = %q, want %q", got, tt.want)
			}
		})
	}
}
