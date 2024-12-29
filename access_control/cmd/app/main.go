package main

import (
	"accessCtf/internal/config"
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	//if err := godotenv.Load("./config/.app_envc"); err != nil {
	//}
	if err := godotenv.Load("./config/.storage_env"); err != nil {
		log.Fatalf("Error with loading StorageEnv file: %v", err)
	}
	cfg := config.Load("./config/local.yaml")
	fmt.Println(*cfg)
}
