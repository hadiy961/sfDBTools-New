package execution

import "testing"

func TestRemoveUnsupportedMysqldumpOption_RemovesMaxStatementTime(t *testing.T) {
	args := []string{
		"--host=127.0.0.1",
		"--user=root",
		"--max-statement-time=0",
		"--single-transaction",
		"db1",
	}
	stderr := "/usr/bin/mysqldump: unknown variable 'max-statement-time=0'\n"

	newArgs, removed, ok := RemoveUnsupportedMysqldumpOption(args, stderr)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if removed != "--max-statement-time=0" {
		t.Fatalf("unexpected removed: %q", removed)
	}
	for _, a := range newArgs {
		if a == "--max-statement-time=0" {
			t.Fatalf("expected --max-statement-time=0 to be removed, args=%v", newArgs)
		}
	}
}

func TestRemoveUnsupportedMysqldumpOption_RemovesAllIgnoreTableDataFlags(t *testing.T) {
	args := []string{
		"--host=127.0.0.1",
		"--user=root",
		"--ignore-table-data=db1.tsflpageview",
		"--ignore-table-data=db1.tsfltokensess",
		"--ignore-table-data=db1.tcllusersession",
		"--single-transaction",
		"db1",
	}
	stderr := "/usr/bin/mysqldump: unknown variable 'ignore-table-data=db1.tsflpageview'\n"

	newArgs, removed, ok := RemoveUnsupportedMysqldumpOption(args, stderr)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if removed == "" {
		t.Fatalf("expected removed to be non-empty")
	}
	for _, a := range newArgs {
		if len(a) >= len("--ignore-table-data") && a[:len("--ignore-table-data")] == "--ignore-table-data" {
			t.Fatalf("expected all --ignore-table-data flags removed, found %q in %v", a, newArgs)
		}
	}
}

func TestRemoveUnsupportedMysqldumpOption_RemovesUnknownOptionToken(t *testing.T) {
	args := []string{
		"--set-gtid-purged=OFF",
		"--single-transaction",
		"db1",
	}
	stderr := "mysqldump: unknown option '--set-gtid-purged=OFF'\n"

	newArgs, removed, ok := RemoveUnsupportedMysqldumpOption(args, stderr)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if removed != "--set-gtid-purged=OFF" {
		t.Fatalf("unexpected removed: %q", removed)
	}
	for _, a := range newArgs {
		if a == "--set-gtid-purged=OFF" {
			t.Fatalf("expected --set-gtid-purged=OFF to be removed, args=%v", newArgs)
		}
	}
}

func TestRemoveUnsupportedMysqldumpOption_NoMatch(t *testing.T) {
	args := []string{"--single-transaction", "db1"}
	stderr := "some other error\n"

	newArgs, removed, ok := RemoveUnsupportedMysqldumpOption(args, stderr)
	if ok {
		t.Fatalf("expected ok=false")
	}
	if removed != "" {
		t.Fatalf("expected removed empty, got %q", removed)
	}
	if len(newArgs) != len(args) {
		t.Fatalf("expected args unchanged")
	}
}
