package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	SQLStorageType      = "sql"
	InmemoryStorageType = "inmemory"
)

type Config struct {
	StorageType string `mapstructure:"STORAGE_TYPE"`
	Host        string `mapstructure:"HOST"`
	HTTPPort    string `mapstructure:"HTTP_PORT"`
	GRPCPort    string `mapstructure:"GRPC_PORT"`
	LogFile     string `mapstructure:"LOG_FILE"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	DBUser      string `mapstructure:"DB_USER"`
	DBPassword  string `mapstructure:"DB_PASSWORD"`
	DBName      string `mapstructure:"DB_NAME"`
	DBPort      string `mapstructure:"DB_PORT"`
	DBHost      string `mapstructure:"DB_HOST"`
}

func NewConfig(configPath string) Config {
	var config Config
	readConfig(configPath)
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	return config
}

func readConfig(configPath string) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
