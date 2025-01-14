package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type AppConfig struct {
	Env           string     `yaml:"env" env:"ENV" env-default:"local"`
	StorageCfg    Storage    `yaml:"storage" env-required:"true"`
	HTTPServerCfg HttpServer `yaml:"httpServer"`
	ImagesCfg     Images     `yaml:"images"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8082"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idleTimeout" env-default:"60s"`
}

type Storage struct {
	Path     string `yaml:"path" env:"DB_PATH" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DbName   string `env:"POSTGRES_DB" env-required:"true"`
}
type Images struct {
	Path        string `yaml:"basePath" env:"IMAGES_BASE_PATH" env-default:"./static/users/upload/"`
	AvatarsPath string `yaml:"avatarsPath" env:"AVATARS_BASE_PATH" env-default:"./static/users/avatars/"`
	//DefaultAdminImageURL string `yaml:"defaultAdminImageURL" env:"DEFAULT_ADMIN_IMAGE_URL" env-default:"/static/users/upload/default/admin.jpg"`
}

func Load(configPath string) *AppConfig {
	const fn = "config:Load"
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("%s: ConfigPath error: %v", fn, err)
	}

	var config AppConfig

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("Config reading error: %v", err)
	}
	return &config
}
