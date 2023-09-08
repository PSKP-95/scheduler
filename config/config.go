package config

import (
	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Driver string `mapstructure:"DB_DRIVER"`
	URL    string `mapstructure:"DB_SOURCE"`
}

type ServerConfig struct {
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

type WorkerConfig struct {
	WorkLookAheadSec string `mapstructure:"WORK_LOOK_AHEAD_SEC"`
}

func LoadConfig(path string) (DatabaseConfig, ServerConfig, WorkerConfig, error) {
	var dbConfig DatabaseConfig
	var serverConfig ServerConfig
	var workerConfig WorkerConfig

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return dbConfig, serverConfig, workerConfig, err
	}

	err = viper.Unmarshal(&dbConfig)
	if err != nil {
		return dbConfig, serverConfig, workerConfig, err
	}

	err = viper.Unmarshal(&serverConfig)
	if err != nil {
		return dbConfig, serverConfig, workerConfig, err
	}

	err = viper.Unmarshal(&workerConfig)
	if err != nil {
		return dbConfig, serverConfig, workerConfig, err
	}

	return dbConfig, serverConfig, workerConfig, nil
}
