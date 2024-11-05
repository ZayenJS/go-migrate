package util

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/ZayenJS/go-migrate/database"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func GetDatabaseURLFromEnvFile(currentWorkingDirectory string) string {
	dotEnvPath := path.Join(currentWorkingDirectory, ".env")
	dotEnvFileContent, err := os.ReadFile(dotEnvPath)

	if err != nil {
		fmt.Printf("Error reading .env file: %v\n", err)
		os.Exit(1)
	}

	regexp, err := regexp.Compile(`((postgres|mysql)://.*)`)

	if err != nil {
		fmt.Printf("Error compiling regex: %v\n", err)
		os.Exit(1)
	}

	matches := regexp.FindStringSubmatch(string(dotEnvFileContent))

	if len(matches) < 1 {
		fmt.Println("Error reading database URL from .env file. Make sure it is in the format GO_MIGRATE_DATABASE_URL=<dialect>://<username>:<password>@<host>:<port>/<database>")
		os.Exit(1)
	}

	return matches[0]
}

func GetCreateMigrationTableQuery() string {
	dialect := database.Get().Dialect

	if dialect == "postgres" {
		return "CREATE TABLE IF NOT EXISTS migrations (id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"
	}

	return "CREATE TABLE IF NOT EXISTS migrations (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255) NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"
}

func GetSelectMigrationPreparedQuery() string {
	dialect := database.Get().Dialect

	if dialect == "postgres" {
		return "SELECT name FROM migrations WHERE name = $1"
	}

	return "SELECT name FROM migrations WHERE name = ?"
}

func GetInsertMigrationPreparedQuery() string {
	dialect := database.Get().Dialect

	if dialect == "postgres" {
		return "INSERT INTO migrations (name) VALUES ($1)"
	}

	return "INSERT INTO migrations (name) VALUES (?)"
}
