package config

import (
	"github.com/caarlos0/env/v10"
)

type Redis struct {
	Address  string `env:"REDIS_ADDRESS,required"`
	Password string `env:"REDIS_PASSWORD,required"`
}

type Gilas struct {
	ApiKey string `env:"GILAS_API_KEY,required"`
}

type Telegram struct {
	ApiKey           string `env:"TELEGRAM_API_KEY,required"`
	GroupId          int64  `env:"TELEGRAM_GROUP_ID,required"`
	MessageRateLimit int    `env:"TELEGRAM_MESSAGE_RATE_LIMIT" envDefault:"2"`
}

type Config struct {
	Redis    Redis
	Gilas    Gilas
	Telegram Telegram
}

func LoadConfigOrPanic() Config {
	var config *Config = new(Config)
	if err := env.Parse(config); err != nil {
		panic(err)
	}

	return *config
}
