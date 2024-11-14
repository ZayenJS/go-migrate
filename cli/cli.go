package cli

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ZayenJS/go-migrate/config"
	"github.com/ZayenJS/go-migrate/database"
	"github.com/ZayenJS/go-migrate/filesystem"
	migration "github.com/ZayenJS/go-migrate/models"
	"github.com/ZayenJS/go-migrate/util"
)

type CLI struct{}

type commands struct {
	Init     string
	Create   string
	Migrate  string
	Rollback string
}

var Commands = commands{
	Init:     "init",
	Create:   "create",
	Migrate:  "migrate",
	Rollback: "rollback",
}

func Setup() {
	flag.Usage = func() {
		fmt.Println("Usage: go-migrate [command]")
		fmt.Println("Migrate database using SQL files in migrations directory")
		fmt.Println("Commands:")
		fmt.Println("init - Initialize migrations directory, create .env file with GO_MIGRATE_DATABASE_URL and config file with directory path")
		fmt.Println("create [name] - Create new migration with name")
		fmt.Println("migrate - Run all pending migrations")
		fmt.Println("rollback [steps] - Rollback last migration or [steps] migrations")
		fmt.Println("Options:")
		fmt.Println("  -h, --help: Show help message")

		os.Exit(0)
	}
	flag.Parse()
}

func Init() {
	currentWorkingDirectory := filesystem.GetCurrentWorkingDirectory()
	configuration := config.Get()
	absoluteDirectoryPath := path.Join(currentWorkingDirectory, configuration.GetMigrationDirectoryName())
	filesystem.CreateDirectoryIfNotExist(absoluteDirectoryPath)

	envFilePath := path.Join(currentWorkingDirectory, ".env")
	filesystem.CreateFileIfNotExist(envFilePath, "GO_MIGRATE_DATABASE_URL=<dialect>://<username>:<password>@<host>:<port>/<database>?multiStatements=true", os.O_APPEND|os.O_CREATE|os.O_WRONLY)

	configFilePath := path.Join(currentWorkingDirectory, configuration.GetConfigFileName())
	filesystem.CreateFileIfNotExist(configFilePath, fmt.Sprintf("{\"directoryPath\": \"%v\"}", absoluteDirectoryPath), os.O_CREATE)

	fmt.Println("Initialization complete")
}

func Create() {
	configuration := config.Get()
	currentWorkingDirectory := filesystem.GetCurrentWorkingDirectory()
	absoluteDirectoryPath := path.Join(currentWorkingDirectory, configuration.GetMigrationDirectoryName())
	filesystem.CreateDirectoryIfNotExist(absoluteDirectoryPath)

	migrationName := flag.Arg(1)
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()

	migrationFilePath := path.Join(
		absoluteDirectoryPath,
		fmt.Sprintf(
			"%v_%v",
			fmt.Sprintf("%v%02d%02d%02d%02d%02d", year, month, day, hour, minute, second),
			migrationName,
		),
	)

	filesystem.CreateDirectoryIfNotExist(migrationFilePath)

	upFilePath := path.Join(migrationFilePath, "up.sql")
	downFilePath := path.Join(migrationFilePath, "down.sql")
	filesystem.CreateFileIfNotExist(upFilePath, "", os.O_CREATE)
	filesystem.CreateFileIfNotExist(downFilePath, "", os.O_CREATE)

	fmt.Println("Migration files created")
}

func Migrate() {
	currentWorkingDirectory := filesystem.GetCurrentWorkingDirectory()

	databaseURL := util.GetDatabaseURLFromEnvFile(currentWorkingDirectory)
	db := database.Connect(databaseURL)

	defer db.Engine.Close()

	database.CreateMigrationTableIfNotExists()

	configuration := config.Get()
	configFileContent := filesystem.GetFileContent(path.Join(currentWorkingDirectory, configuration.GetConfigFileName()))
	json.Unmarshal([]byte(configFileContent), configuration)

	migrationApplied := 0

	migrationDirectoryContent := filesystem.GetDirectoryContent(configuration.DirectoryPath)

	for _, file := range migrationDirectoryContent {
		entryName := file.Name()
		migrationStruct := migration.GetByName(entryName)

		if migrationStruct.Name != "" {
			continue
		}

		migrationFilePath := path.Join(configuration.DirectoryPath, entryName, "up.sql")
		sqlText := filesystem.GetFileContent(migrationFilePath)

		fmt.Printf("Running migration file %v\n\n", entryName)
		time.Sleep(500 * time.Millisecond)

		fmt.Println(sqlText)

		executeInTransaction(db, func(tx *sql.Tx) {
			_, err := tx.Exec(sqlText)

			if err != nil {
				fmt.Printf("\n\033[31mError running migration file %v:\n\033[0m\n%v\n\n", file.Name(), err)
				os.Exit(1)
			}
		})

		migration.Insert(entryName)

		fmt.Println("\033[32mMigration complete\033[0m")
		fmt.Print("\n")
		migrationApplied++

		time.Sleep(500 * time.Millisecond)
	}

	if migrationApplied == 0 {
		fmt.Println("No migrations to run")
		os.Exit(0)
	}

	fmt.Println("All migrations complete")
}

func Rollback(steps int64) {
	currentWorkingDirectory := filesystem.GetCurrentWorkingDirectory()

	databaseURL := util.GetDatabaseURLFromEnvFile(currentWorkingDirectory)
	db := database.Connect(databaseURL)

	sqlText := "SELECT * FROM migrations ORDER BY id DESC LIMIT ?"
	if db.Dialect == database.Dialect.Postgres {
		sqlText = "SELECT * FROM migrations ORDER BY id DESC LIMIT $1"
	}

	defer db.Engine.Close()

	if steps == 0 {
		steps = 1
	}

	rows, err := db.Engine.Query(sqlText, steps)

	if err != nil {
		fmt.Printf("Error rolling back migration: %v\n", err)
		os.Exit(1)
	}

	migrationsToRollback := make([]string, 0)

	for rows.Next() {
		migration := migration.New()
		err := rows.Scan(&migration.ID, &migration.Name, &migration.CreatedAt)

		if err != nil {
			fmt.Printf("Error scanning migration: %v\n", err)
			os.Exit(1)
		}

		if migration.Name == "" {
			continue
		}

		migrationsToRollback = append(migrationsToRollback, migration.Name)
	}

	if len(migrationsToRollback) == 0 {
		fmt.Println("Nothing to rollback")
		os.Exit(0)
	}

	text := "Rolling back %v migration"

	if len(migrationsToRollback) > 1 {
		text += "s"
	}

	fmt.Printf(text+"\n", len(migrationsToRollback))

	configuration := config.Get()
	configFileContent := filesystem.GetFileContent(path.Join(currentWorkingDirectory, configuration.GetConfigFileName()))
	json.Unmarshal([]byte(configFileContent), configuration)

	migrationDirectoryContent := filesystem.GetDirectoryContent(configuration.DirectoryPath)

	rolledBackMigrations := 0

	for _, migrationName := range migrationsToRollback {
		fmt.Printf("\nRolling back migration %v\n", migrationName)
		time.Sleep(500 * time.Millisecond)

		for _, file := range migrationDirectoryContent {
			if file.Name() != migrationName {
				continue
			}

			migrationFilePath := path.Join(configuration.DirectoryPath, migrationName, "down.sql")
			sqlText := filesystem.GetFileContent(migrationFilePath)

			if sqlText == "" {
				fmt.Printf("\033[33mThe down.sql file for migration %v is either empty or missing\033[0m\n", migrationName)

				migration.Delete(migrationName)

				if steps > 1 {
					fmt.Println("Continuing to next migration")
				} else {
					fmt.Println("\033[32mRollback complete\033[0m")
				}

				continue
			}

			fmt.Println(sqlText)

			executeInTransaction(db, func(tx *sql.Tx) {
				_, err := tx.Exec(sqlText)

				if err != nil {
					fmt.Printf("Error rolling back migration file %v:\n%v\n\n", migrationName, err)
					os.Exit(1)
				}
			})

			time.Sleep(500 * time.Millisecond)
		}

		migration.Delete(migrationName)

		fmt.Println("\033[32mMigration rolled back\033[0m")
		fmt.Print("\n")
		rolledBackMigrations++
	}

	if rolledBackMigrations > 0 {
		text = "Rolled back %v migration"

		if rolledBackMigrations > 1 {
			text += "s"
		}

		text += "\n"

		fmt.Printf(text, rolledBackMigrations)
	}
}

func executeInTransaction(db *database.DataBase, callback func(*sql.Tx)) {
	tx, err := db.Engine.Begin()

	if err != nil {
		fmt.Errorf("Error starting transaction: %v\n", err)
		os.Exit(1)
	}

	defer tx.Rollback()
	callback(tx)
	tx.Commit()
}
