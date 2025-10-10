package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/verify/main.go <driver> <dsn>")
		fmt.Println("Examples:")
		fmt.Println("  sqlite3 ./expense.db")
		fmt.Println("  postgres \"postgres://user:pass@localhost:5432/expense?sslmode=disable\"")
		os.Exit(1)
	}

	driver := os.Args[1]
	dsn := os.Args[2]

	if driver == "sqlite3" && strings.HasPrefix(dsn, "sqlite3://") {
		dsn = strings.TrimPrefix(dsn, "sqlite3://")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v\n", err)
	}
	defer db.Close()

	if driver == "sqlite3" {
		_, _ = db.Exec("PRAGMA foreign_keys = ON;")
	}

	tables := []string{"users", "categories", "expenses"}
	for _, t := range tables {
		exists, err := tableExists(db, driver, t)
		if err != nil {
			fmt.Printf("%s: error checking existence: %v\n", t, err)
			continue
		}
		if !exists {
			fmt.Printf("%s: NOT FOUND\n", t)
			continue
		}
		var cnt int
		row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", t))
		if err := row.Scan(&cnt); err != nil {
			fmt.Printf("%s: error counting rows: %v\n", t, err)
			continue
		}
		fmt.Printf("%s: OK (rows=%d)\n", t, cnt)
	}
}

func tableExists(db *sql.DB, driver, table string) (bool, error) {
	switch driver {
	case "sqlite3":
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?;", table).Scan(&count)
		return count > 0, err
	case "postgres":
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1);", table).Scan(&exists)
		return exists, err
	default:
		return false, fmt.Errorf("unsupported driver: %s", driver)
	}
}
