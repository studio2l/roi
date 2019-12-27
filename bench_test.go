package roi

import "testing"

func BenchmarkDBExecStmt(b *testing.B) {
	db, err := testDB()
	if err != nil {
		b.Fatalf("could not open db: %w", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS bench (val INT)")
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.Exec("TRUNCATE TABLE bench")
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		_, err := db.Exec("INSERT INTO bench (val) VALUES ($1)", i)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDBExecFunc(b *testing.B) {
	db, err := testDB()
	if err != nil {
		b.Fatalf("could not open db: %w", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS bench (val INT)")
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.Exec("TRUNCATE TABLE bench")
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		err := dbExec(db, []dbStatement{dbStmt("INSERT INTO bench (val) VALUES ($1)", i)})
		if err != nil {
			b.Fatal(err)
		}
	}
}
