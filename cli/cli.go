package cli

import (
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
		fmt.Println("init - Initialize migrations directory, create .env file with DATABASE_URL and config file with directory path")
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
	filesystem.CreateFileIfNotExist(envFilePath, "DATABASE_URL=<dialect>://<username>:<password>@<host>:<port>/<database>")

	configFilePath := path.Join(currentWorkingDirectory, configuration.GetConfigFileName())
	filesystem.CreateFileIfNotExist(configFilePath, fmt.Sprintf("{\"directoryPath\": \"%v\"}", absoluteDirectoryPath))

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
	filesystem.CreateFileIfNotExist(upFilePath, "")
	filesystem.CreateFileIfNotExist(downFilePath, "")

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
		fmt.Println(string(sqlText))

		_, err := db.Engine.Exec(string(sqlText))

		if err != nil {
			fmt.Printf("Error running migration file %v:\n%v\n\n", file.Name(), err)
			os.Exit(1)
		}

		migration.Insert(entryName)

		fmt.Println("Migration complete")
		fmt.Print("\n")
		migrationApplied++
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

	sql := "SELECT * FROM migrations ORDER BY id DESC LIMIT ?"
	if db.Dialect == database.Dialect.Postgres {
		sql = "SELECT * FROM migrations ORDER BY id DESC LIMIT $1"
	}

	defer db.Engine.Close()

	if steps == 0 {
		steps = 1
	}

	rows, err := db.Engine.Query(sql, steps)

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
		fmt.Printf("Rolling back migration %v\n", migrationName)

		for _, file := range migrationDirectoryContent {
			if file.Name() != migrationName {
				continue
			}

			migrationFilePath := path.Join(configuration.DirectoryPath, migrationName, "down.sql")
			sqlText := filesystem.GetFileContent(migrationFilePath)

			fmt.Println(string(sqlText))

			_, err := db.Engine.Exec(string(sqlText))

			if err != nil {
				fmt.Printf("Error rolling back migration file %v:\n%v\n\n", migrationName, err)
				os.Exit(1)
			}

			migration.Delete(migrationName)

			fmt.Println("Migration rolled back")
			fmt.Print("\n")
			rolledBackMigrations++
		}
	}

	if rolledBackMigrations > 0 {
		text = "Rolled back %v migration"

		if rolledBackMigrations > 1 {
			text += "s"
		}

		fmt.Printf(text, rolledBackMigrations)
	}
}
