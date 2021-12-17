package config

import (
	"log"

	"github.com/spf13/viper"
)

var config Config

type Config struct {
	WebsocketURL          string `mapstructure:"WEBSOCKET_URL"`
	RPCURL                string `mapstructure:"RPC_URL"`
	MongoURI              string `mapstructure:"MONGO_URI"`
	MongoDBName           string `mapstructure:"MONGO_DB_NAME"`
	Concurrency           int    `mapstructure:"CONCURRENCY"`
	NumberOfConfirmations uint64 `mapstructure:"NUMBER_OF_CONFIRMATIONS"`
	MaxJobTimeout         int    `mapstructure:"MAX_JOB_TIMEOUT"`
}

func Read(file string) {
	viper.SetDefault("MONGO_DB_NAME", "evm-indexer")
	viper.SetDefault("MAX_JOB_TIMEOUT", 5)
	viper.SetConfigFile(file)
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("❌ failed to read `.env` file : %s\n", err.Error())
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("❌ failed to unmarshals the config into a struct : %s\n", err.Error())
	}
}

func Get() Config {
	return config
}
