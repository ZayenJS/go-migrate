package database

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type DataBase struct {
	Engine  *sql.DB
	Dialect string
}

type dialect struct {
	Mysql    string
	Postgres string
}

var Dialect = dialect{
	Mysql:    "mysql",
	Postgres: "postgres",
}

var dbInstance = DataBase{}

func Get() *DataBase {
	return &dbInstance
}

func Connect(databaseURL string) *DataBase {
	re := regexp.MustCompile(`(postgres|mysql)`)
	dialect := re.FindString(databaseURL)

	var db *sql.DB
	var err error

	if dialect == "postgres" {
		db, err = sql.Open(dialect, databaseURL)
	} else if dialect == "mysql" {
		dsn := strings.Replace(strings.Split(databaseURL, "://")[1], `"`, "", -1)
		db, err = sql.Open(dialect, dsn)
	} else {
		fmt.Println("Invalid dialect")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	dbInstance.Engine = db
	dbInstance.Dialect = dialect

	return &dbInstance
}

func CreateMigrationTableIfNotExists() {
	sql := "CREATE TABLE IF NOT EXISTS migrations (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255) NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"

	if dbInstance.Dialect == "postgres" {
		sql = "CREATE TABLE IF NOT EXISTS migrations (id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"
	}

	_, err := dbInstance.Engine.Query(sql)

	if err != nil {
		fmt.Printf("Error creating migrations table: %v\n", err)
		os.Exit(1)
	}
}
