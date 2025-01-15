package main

import (
	"accessCtf/internal/config"
	"accessCtf/internal/storage"
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"log"
	"os"
	"time"
)

import (
	_ "accessCtf/migrations"
)

func main() {

	migrationsDirPath := flag.String("mgPath", "./migrations/", "path to folder with goose migratons")
	storagePath := flag.String("dbPath", "localhost:54321", "path to database")
	configEnvPath := flag.String(
		"envPath",
		"./config/.storage_env",
		"path to env config file with 3 defined keys. POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB",
	)
	isUp := flag.Bool(
		"up",
		false,
		"Apply all available migrations. Default: false.",
	)
	downgrade := flag.Int64(
		"down-to",
		-1,
		"downgrade database by {{n}} migrations. Default: false.",
	)
	isDowngradeAll := flag.Bool(
		"down-all",
		false,
		"downgrade ALL MIGRATIONS. Shortcut for \"-down-to 0\" Default: false.",
	)
	flag.Parse()

	fi, err := os.Stat(*migrationsDirPath)
	if err != nil {
		log.Fatalf("Error opening config env file: %v", err)
	}
	if !fi.IsDir() {
		log.Fatalf("Path %s not point to directory", *migrationsDirPath)
	}
	if err := os.Setenv("DB_PATH", *storagePath); err != nil {
		log.Fatalf("Error with setting DB_PATH env. variable: %v", err)
	}
	if _, err := os.Stat(*configEnvPath); err != nil {
		log.Fatalf("Error opening config env file: %v", err)
	}
	if err := godotenv.Load(*configEnvPath); err != nil {
		log.Fatalf("Error with loading env vars from config file: %v", err)
	}
	var strgCfg config.Storage
	err = cleanenv.ReadConfig("./config/local.yaml", &strgCfg)
	if err != nil {
		log.Fatalf("MIGRATOR: Error reading config file: %v", err)
	}
	pool, err := storage.NewPgStorage(strgCfg, time.Second*2)
	if err != nil {
		log.Fatalf("MIGRATOR: Error connecting to database: %v", err)
	}
	db := stdlib.OpenDBFromPool(pool.Conn)

	if err := godotenv.Load("./config/.app_env"); err != nil {
		log.Fatalf("Error with loading AppEnv file: %v", err)
	}

	if *isUp {
		if err := goose.Up(db, *migrationsDirPath); err != nil {
			log.Fatalf("Error with migrations.Up: %v", err)
		}
		return
	}
	if *downgrade >= 0 {
		if err := goose.DownTo(db, *migrationsDirPath, *downgrade); err != nil {
			log.Fatalf("Error with migrations.DownTo %d: %v", *downgrade, err)
		}
		return
	}
	if *isDowngradeAll {
		if err := goose.DownTo(db, *migrationsDirPath, 0); err != nil {
			log.Fatalf("Error with migrations.DownAll: %v", err)
		}
		return
	}
	// default case. show versions
	if err := goose.Status(db, *migrationsDirPath); err != nil {
		log.Fatalf("Error with migrations.Status: %v", err)
	}
}
