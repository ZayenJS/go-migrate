package config

type Config struct {
	DirectoryPath string `json:"directoryPath"`
}

var ConfigInstance Config = Config{}

func Get() *Config {
	return &ConfigInstance
}

func (c *Config) GetConfigFileName() string {
	return "go-migrate.config.json"
}

func (c *Config) GetMigrationDirectoryName() string {
	return "migrations"
}
