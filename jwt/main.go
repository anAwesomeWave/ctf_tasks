package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	secret = []byte("mysecret")
)

func GetJwtToken(r *http.Request) string {
	//from header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	//from cookie
	for _, cookie := range r.Cookies() {
		if cookie.Name == "jwt" {
			return cookie.Value
		}
	}

	return ""
}

func VulnerableValidate(r *http.Request) (bool, error) {
	tokenStr := GetJwtToken(r)
	if tokenStr == "" {
		return false, fmt.Errorf("GetJwtToken: no token found")
	}
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return false, fmt.Errorf("Ошибка при парсинге токена: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		fmt.Println("Token is NOT valid")
		return false, nil
	}

	fmt.Println("Token is  valid:", claims)

	userString, ok := claims["user"].(string)
	return ok && userString == "admin", nil
}

func sendHttpFlag(w http.ResponseWriter, r *http.Request) {
	if ok, err := VulnerableValidate(r); err != nil || !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		fmt.Println("VulnerableValidate - return False", ok, err)
		return
	}
	w.Write([]byte(os.Getenv("CTF_FLAG")))
}

func main() {
	// Генерируем токен с алгоритмом HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "alice",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Сгенерированный токен:", tokenStr)

	http.HandleFunc("/", sendHttpFlag)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe("0.0.0.0:8081", nil))
}
