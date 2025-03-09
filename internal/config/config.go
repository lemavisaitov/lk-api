package config

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	APIPort       string        `mapstructure:"API_ADDRESS"`
	APIHost       string        `mapstructure:"API_ADDRESS"`
	DBUser        string        `mapstructure:"DB_USER"`
	DBPassword    string        `mapstructure:"DB_PASSWORD"`
	DBHost        string        `mapstructure:"DB_HOST"`
	DBPort        string        `mapstructure:"DB_PORT"`
	DBConnTimeout time.Duration `mapstructure:"DB_CONN_TIMEOUT"`
	CacheInterval time.Duration `mapstructure:"CACHE_INTERVAL"`
	CacheTTL      time.Duration `mapstructure:"CACHE_TTL"`
}

func MustLoad() *Config {
	cfg := Config{}
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(errors.Wrap(err, "viper.ReadInConfig"))
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(errors.Wrap(err, "viper.Unmarshal"))
	}

	return &cfg
}
