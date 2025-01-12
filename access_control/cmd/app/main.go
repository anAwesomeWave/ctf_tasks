package main

import (
	"accessCtf/internal/app"
	"accessCtf/internal/config"
	"accessCtf/internal/storage"
	"context"
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
	pgCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	pgStrg, err := storage.NewPgStorage(cfg.StorageCfg, pgCtx)

	if err != nil {
		log.Fatal(err)
	}

	uuid, err := pgStrg.CreateUser("timus", "77777Tim")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(uuid.String())
	defaultApp, err := app.NewDefaultApp(cfg.ImagesCfg.Path, cfg.ImagesCfg.AvatarsPath)
	if err != nil {
		log.Fatal(err)
	}
	router := setUpRouter(defaultApp)

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
