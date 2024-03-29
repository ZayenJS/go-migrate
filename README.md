# GO Migrate

This is a simple tool to manage your database migrations. It is written in Go and uses simple commands to manage your database migrations.

## Installation

To install the tool, you can use the following command:

```bash
go get github.com/zayenjs/go-migrate
```

## Usage

To use the tool, you can use the following commands:

```bash
go-migrate init
go-migrate create <migration_name>
go-migrate migrate
go-migrate rollback [steps]
```

## Configuration

The tool uses a configuration file to get some configurations. The config file can be created using the `init` command.

The configuration file is a simple JSON file that contains the following fields:

```json
{
  "directoryPath": "/home/xxx/xxx/migrations"
}
```

The `directoryPath` field is the path to the directory where the migration files will be stored.

## Commands

### Init

The `init` command will create the following:

- The configuration file.
- The migrations directory.
- The migrations table in the database.
- The .env file (if it does not exist).

> The .env file is used to store the database connection string. The file should have a `DATABASE_URL` field that contains the connection string otherwise the tool will not work.

### Create

The `create` command will create a new migration file in the migrations directory. The command takes one argument which is the migration name.

### Migrate

The `migrate` command will run all the pending migrations.

### Rollback

The `rollback` command will rollback the last migration. You can also rollback multiple migrations by passing the number of migrations to rollback.
If the number of migrations is not passed, the command will rollback the last migration.
