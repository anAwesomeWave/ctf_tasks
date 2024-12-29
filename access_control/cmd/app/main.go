package main

import (
	"accessCtf/internal/config"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	//if err := godotenv.Load("./config/.app_envc"); err != nil {
	//}
	if err := godotenv.Load("./config/.storage_env"); err != nil {
		log.Fatalf("Error with loading StorageEnv file: %v", err)
	}
	cfg := config.Load("./config/local.yaml")
	fmt.Println(*cfg)

	router := setUpRouter()

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
