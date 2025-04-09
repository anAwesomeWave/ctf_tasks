package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

)


func main() {
	port := 8081

	initDB()

	// http.HandleFunc("/register", registerHandler)
	// http.HandleFunc("/bonus", bonusHandler)
	// http.HandleFunc("/flag", flagHandler)

	log.Println("Server started on :%d", port)
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
