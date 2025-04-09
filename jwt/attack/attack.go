package main

import (
	"fmt"
	"os"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func main() {
	publicKey, _ := os.ReadFile("pub.txt")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "admin",
		"exp":  time.Now().Add(time.Hour * 1).Unix(),
	})
	

	log.Println(publicKey)

	tokenString, err := token.SignedString(publicKey)
	if err != nil {
		fmt.Println("Ошибка при создании токена:", err)
		return
	}

	fmt.Println("HS256 токен:", tokenString)
}
