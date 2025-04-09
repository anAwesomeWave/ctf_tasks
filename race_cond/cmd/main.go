package main

import (
	"fmt"
	"log"
	"net/http"
	"race_cond/internal/storage"
)

func main() {
	port := 8081

	strg := storage.InitDB()

	// http.HandleFunc("/register", registerHandler)
	// http.HandleFunc("/bonus", bonusHandler)
	// http.HandleFunc("/flag", flagHandler)
	router := setUpRouter(strg)
	log.Printf("Server started on :%d\n", port)
	serv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: router,
	}
	if err := serv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
