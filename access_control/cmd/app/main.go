package main

import (
	"accessCtf/internal/app"
	"accessCtf/internal/config"
	"accessCtf/internal/storage"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"time"
)

func main() {
	//if err := godotenv.Load("./config/.app_envc"); err != nil {
	//}
	if err := godotenv.Load("./config/.storage_env"); err != nil {
		log.Fatalf("Error with loading StorageEnv file: %v", err)
	}
	cfg := config.Load("./config/local.yaml")

	fmt.Println(*cfg)
	pgStrg, err := storage.NewPgStorage(cfg.StorageCfg, time.Millisecond*500)

	if err != nil {
		log.Fatal(err)
	}

	defaultApp, err := app.NewDefaultApp(cfg.ImagesCfg.Path, cfg.ImagesCfg.AvatarsPath)
	if err != nil {
		log.Fatal(err)
	}
	router := setUpRouter(defaultApp, pgStrg)

	serv := &http.Server{
		Addr:         cfg.HTTPServerCfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServerCfg.Timeout,
		WriteTimeout: cfg.HTTPServerCfg.Timeout,
		IdleTimeout:  cfg.HTTPServerCfg.IdleTimeout,
	}
	if err := serv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
