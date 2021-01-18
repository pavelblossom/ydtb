package config

import (
	"github.com/caarlos0/env"
)

type Config struct {
	TelegramToken string `env:"TELEGRAM_TOKEN,required"`
	YoutubeToken  string `env:"YOUTUBE_TOKEN,required"`
	ProxyURL      string `env:"PROXY_URL"`
	UseProxy      bool   `env:"USE_PROXY" envDefault:"false"`
	Concurrency   int    `env:"CONCURRENCY" envDefault:"100"`
	DownloadsDir  string `env:"DOWNLOADS_DIR" envDefault:"files"`
}

func Get() (*Config, error) {
	var c Config
	err := env.Parse(&c)
	return &c, err
}
