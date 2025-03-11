package config

import (
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppAddress    string        `env:"APP_ADDRESS" env-default:":8080"`
	DBUser        string        `env:"DB_USER" env-default:"postgres"`
	DBPassword    string        `env:"DB_PASSWORD" env-default:"postgres"`
	DBHost        string        `env:"DB_HOST" env-default:"postgres"`
	DBPort        string        `env:"DB_PORT" env-default:"5432"`
	DBConnTimeout time.Duration `env:"DB_CONN_TIMEOUT" env-default:"5s"`
	CacheInterval time.Duration `env:"CACHE_INTERVAL" env-default:"5s"`
	CacheTTL      time.Duration `env:"CACHE_TTL" env-default:"10s"`
}

func Load() (*Config, error) {
	cfg := Config{}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to read env")
	}

	return &cfg, nil
}

func (cfg *Config) GetDBConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
}
