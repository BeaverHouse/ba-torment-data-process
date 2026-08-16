package party

import (
	"database/sql"
	"path/filepath"
	"testing"

	"ba-torment-data-process/internal/logic/id"
)

func TestGetPartiesExtractsStartingSkillOrder(t *testing.T) {
	db := newPartyTestDB(t, true)
	execPartyTestSQL(t, db, `INSERT INTO runs VALUES (74, 98271921), (73, 98271921)`)
	execPartyTestSQL(t, db, `INSERT INTO students VALUES
		(10089, 'UE40', 90, 0, false, 73, 2),
		(10128, 'UE40', 90, 1, false, 73, 3),
		(10148, 'UE60', 90, 2, true, 73, 5),
		(10105, 'UE40', 90, 3, false, 73, 1),
		(20041, 'UE60', 90, 4, false, 73, 4),
		(20039, 'UE60', 90, 5, false, 73, 0),
		(10145, 'UE40', 90, 0, false, 74, 4),
		(20053, 'UE40', 90, 5, false, 74, 1)`)

	hasSkillOrder, err := hasMulliganColumn(db, "")
	if err != nil {
		t.Fatalf("check mulligan column: %v", err)
	}
	parties, skillOrders, err := getPartiesByCompleteRunID(db, "", 98271921, hasSkillOrder)
	if err != nil {
		t.Fatalf("get parties: %v", err)
	}

	if len(parties) != 2 || len(skillOrders) != 2 {
		t.Fatalf("party counts = %d/%d, want 2/2", len(parties), len(skillOrders))
	}
	wantOrders := [][6]int{{2, 3, 5, 1, 4, 0}, {4, 0, 0, 0, 1, 0}}
	for i := range wantOrders {
		if skillOrders[i] != wantOrders[i] {
			t.Fatalf("party %d skill order = %v, want %v", i+1, skillOrders[i], wantOrders[i])
		}
	}
	wantAssist := id.ComposeStudentDetailID(10148, 5, 4, true)
	if parties[0][2] != wantAssist {
		t.Fatalf("assist slot = %d, want %d", parties[0][2], wantAssist)
	}
}

func TestGetPartiesSupportsLegacySchemaWithoutStartingSkillOrder(t *testing.T) {
	db := newPartyTestDB(t, false)
	execPartyTestSQL(t, db, `INSERT INTO runs VALUES (1, 10)`)
	execPartyTestSQL(t, db, `INSERT INTO students VALUES (10089, 'UE40', 90, 0, false, 1)`)

	hasSkillOrder, err := hasMulliganColumn(db, "")
	if err != nil {
		t.Fatalf("check mulligan column: %v", err)
	}
	if hasSkillOrder {
		t.Fatal("legacy students table unexpectedly has mulligan column")
	}
	parties, skillOrders, err := getPartiesByCompleteRunID(db, "", 10, hasSkillOrder)
	if err != nil {
		t.Fatalf("get legacy parties: %v", err)
	}

	if len(parties) != 1 {
		t.Fatalf("party count = %d, want 1", len(parties))
	}
	if skillOrders != nil {
		t.Fatalf("legacy skill orders = %v, want nil", skillOrders)
	}
}

func TestGetPartiesOmitsEmptyStartingSkillOrder(t *testing.T) {
	db := newPartyTestDB(t, true)
	execPartyTestSQL(t, db, `INSERT INTO runs VALUES (1, 10)`)
	execPartyTestSQL(t, db, `INSERT INTO students VALUES (10089, 'UE40', 90, 0, false, 1, 0)`)

	_, skillOrders, err := getPartiesByCompleteRunID(db, "", 10, true)
	if err != nil {
		t.Fatalf("get parties: %v", err)
	}
	if skillOrders != nil {
		t.Fatalf("empty skill orders = %v, want nil", skillOrders)
	}
}

func newPartyTestDB(t *testing.T, withMulligan bool) *sql.DB {
	t.Helper()
	db, err := sql.Open("duckdb", filepath.Join(t.TempDir(), "party.db"))
	if err != nil {
		t.Fatalf("open duckdb: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close duckdb: %v", err)
		}
	})

	execPartyTestSQL(t, db, `CREATE TABLE runs (runid UINTEGER, crunid UBIGINT)`)
	columns := `sid INTEGER, build VARCHAR, level UTINYINT, slot UTINYINT, assist BOOLEAN, runid UINTEGER`
	if withMulligan {
		columns += `, mulligan UTINYINT`
	}
	execPartyTestSQL(t, db, `CREATE TABLE students (`+columns+`)`)
	return db
}

func execPartyTestSQL(t *testing.T, db *sql.DB, query string) {
	t.Helper()
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("execute fixture SQL: %v", err)
	}
}
