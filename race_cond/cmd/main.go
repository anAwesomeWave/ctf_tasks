package main

import (
	"fmt"
	"github.com/go-chi/jwtauth"
	"log"
	"net/http"
	"os"
	"race_cond/internal/http/handlers"
	"race_cond/internal/storage"
)

func main() {
	port := 8081
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "aboba"
	}
	handlers.TokenAuth = jwtauth.New("HS256", []byte(secretKey), nil)
	strg := storage.InitDB()

	flag := os.Getenv("CTF_FLAG")
	if flag == "" {
		flag = "practice{anawesomewave}"
	}
	router := setUpRouter(strg, flag)
	log.Printf("Server started on :%d\n", port)
	serv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: router,
	}
	if err := serv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
