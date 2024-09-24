package config

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Cookie       *CookieConfig  `env:", prefix=COOKIE_"`
	Passkey      *PasskeyConfig `env:", prefix=PASSKEY_"`
	Server       *ServerConfig  `env:", prefix=SERVER_"`
	User         *UserConfig    `env:", prefix=USER_"`
	EncyptionKey string         `env:"ENCRYPTION_KEY, required"`
}

type ServerConfig struct {
	Port     string `env:"PORT, default=8080"`
	Domain   string `env:"DOMAIN, default=localhost"`
	Protocol string `env:"PROTOCOL, default=http"`
}

type PasskeyConfig struct {
	DisplayName string `env:"DISPLAY_NAME, default=Passkey"`
}

type UserConfig struct {
	ID          string `env:"ID, required"`
	DisplayName string `env:"DISPLAY_NAME, required"`
	Name        string `env:"NAME, required"`
}

type CookieConfig struct {
	Name   string `env:"NAME, default=passkey-backend"`
	Secret string `env:"SECRET, required"`
}

var (
	config Config
	once   sync.Once
)

func LoadConfig() Config {
	once.Do(func() {
		ctx := context.Background()
		log.Print(os.Getenv("USER_ID"))
		if err := envconfig.Process(ctx, &config); err != nil {
			log.Fatal("Failed to process envconfig ", err)
		}
	})
	return config
}
