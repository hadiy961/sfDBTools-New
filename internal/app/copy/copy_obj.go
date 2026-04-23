package copy

import (
	"context"
	"fmt"
	"strings"
	"regexp"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/domain"
)

// copyIndividualObject menyalin satu objek (Procedure/Function/Event) menggunakan SHOW CREATE dan Exec.
func (s *Service) copyIndividualObject(ctx context.Context, client *database.Client, profile *domain.ProfileInfo, sourceDB, targetDB, objType, name string) error {
	query := fmt.Sprintf("SHOW CREATE %s `%s`.`%s` ", objType, sourceDB, name)
	row := client.DB().QueryRowContext(ctx, query)
	
	var n, sql, charSet, collation string
	var err error
	if objType == "EVENT" {
		var sqlMode, tz string
		err = row.Scan(&n, &sqlMode, &tz, &sql, &charSet, &collation, new(interface{}))
	} else {
		err = row.Scan(&n, new(interface{}), &sql, &charSet, &collation, new(interface{}))
	}
	
	if err != nil { return err }

	sql = strings.ReplaceAll(sql, "`"+sourceDB+"`.", "`"+targetDB+"`.")
	sql = strings.ReplaceAll(sql, " "+sourceDB+".", " "+targetDB+".")
	
	re := regexp.MustCompile(`(?i)DEFINER=\s*` + "`" + `.*?` + "`" + `@` + "`" + `.*?` + "`")
	sql = re.ReplaceAllString(sql, "")

	useQuery := fmt.Sprintf("USE `%s` ", targetDB)
	if _, err := client.ExecContextWithRetry(ctx, useQuery); err != nil { return err }
	_, err = client.ExecContextWithRetry(ctx, sql)
	return err
}
