package migration

import (
	"fmt"
	"os"

	"github.com/zayenjs/go-migrate/database"
)

type Migration struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func New() *Migration {
	return &Migration{}
}

func GetByName(name string) *Migration {
	migration := New()
	db := database.Get()
	// check if filename exists in the migration table
	sql := "SELECT * FROM migrations WHERE name = ?"

	if db.Dialect == "postgres" {
		sql = "SELECT * FROM migrations WHERE name = $1"
	}

	stmt, err := db.Engine.Prepare(sql)

	if err != nil {
		fmt.Printf("Error checking if migration exists: %v\n", err)
		os.Exit(1)
	}

	defer stmt.Close()

	err = stmt.QueryRow(name).Scan(&migration.ID, &migration.Name, &migration.CreatedAt)

	if err != nil && err.Error() != "sql: no rows in result set" {
		return nil
	}

	return migration
}

func Insert(name string) {
	db := database.Get()
	sql := "INSERT INTO migrations (name) VALUES (?)"

	if db.Dialect == "postgres" {
		sql = "INSERT INTO migrations (name) VALUES ($1)"
	}

	stmt, err := db.Engine.Prepare(sql)

	if err != nil {
		fmt.Printf("Error preparing insert statement: %v\n", err)
		os.Exit(1)
	}

	defer stmt.Close()

	_, err = stmt.Exec(name)

	if err != nil {
		fmt.Printf("Error inserting migration into migrations table: %v\n", err)
		os.Exit(1)
	}
}

func Delete(name string) {
	db := database.Get()
	sql := "DELETE FROM migrations WHERE name = ?"

	if db.Dialect == "postgres" {
		sql = "DELETE FROM migrations WHERE name = $1"
	}

	stmt, err := db.Engine.Prepare(sql)

	if err != nil {
		fmt.Printf("Error preparing delete statement: %v\n", err)
		os.Exit(1)
	}

	defer stmt.Close()

	_, err = stmt.Exec(name)

	if err != nil {
		fmt.Printf("Error deleting migration from migrations table: %v\n", err)
		os.Exit(1)
	}
}
