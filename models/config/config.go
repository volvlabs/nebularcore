package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Type     types.DatabaseType `yaml:"type" envconfig:"DATABASE_TYPE"`
	Host     string             `yaml:"host" envconfig:"DATABASE_HOST"`
	Username string             `yaml:"username" envconfig:"DATABASE_USERNAME"`
	Password string             `yaml:"password" envconfig:"DATABASE_PASSWORD"`
	Name     string             `yaml:"name" envconfig:"DATABASE_NAME"`
	Port     string             `yaml:"port" envconfig:"DATABASE_PORT"`
	SSLMode  string             `yaml:"sslmode" envconfig:"DATABASE_SSLMODE"`
}

type ServeConfig struct {
	Port            string `yaml:"port" envconfig:"PORT"`
	Host            string `yaml:"host" envconfig:"HOST"`
	ShutdownTimeout int    `yaml:"shutdownTimeout" envconfig:"SHUTDOWN_TIMEOUT"`
	AllowedOrigins  string `yaml:"allowedOrigins" envconfig:"ALLOWED_ORIGINS"`
}

type Endpoints struct {
	AuthEnabled bool `yaml:"authEnabled" envconfig:"AUTH_ENABLED"`
}

type AppConfig struct {
	Env           string         `yaml:"env" envconfig:"ENV"`
	IsDev         bool           `yaml:"isDev" envconfig:"IS_DEV"`
	BaseDir       string         `yaml:"baseDir" envconfig:"BASE_DIR"`
	TestDir       string         `yaml:"testDir" envconfig:"TEST_DIR"`
	EnforceAcl    bool           `yaml:"enforceAcl" envconfig:"ENFORCE_ACL"`
	AutoMigrate   bool           `yaml:"autoMigrate" envconfig:"AUTOMIRGATE"`
	MigrationsDir string         `yaml:"migrationsDir" envconfig:"MIGRATION_DIR"`
	Database      DatabaseConfig `yaml:"database"`
	Server        ServeConfig    `yaml:"server"`
	Endpoints     Endpoints      `yaml:"endpoints"`
}

func readConfigFromFile(configFilePath string) (*AppConfig, error) {
	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	appConfig := &AppConfig{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(appConfig)
	return appConfig, err
}

func New(configFilePath string) (*AppConfig, error) {
	appConfig := &AppConfig{}
	if configFilePath != "" {
		var err error
		appConfig, err = readConfigFromFile(configFilePath)
		if err != nil {
			return nil, err
		}
	}

	envconfig.MustProcess("", appConfig)
	return appConfig, nil
}
