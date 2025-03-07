package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"html/template"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	publicKey      *rsa.PublicKey
	privateKey     *rsa.PrivateKey
	publicKeyBytes []byte
)

func init() {
	// Генерация RSA ключей
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privateKey.PublicKey

	pubASN1, _ := x509.MarshalPKIXPublicKey(publicKey)
	publicKeyBytes = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
}

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
	// from query
	return r.URL.Query().Get("token")
}

func GenerateJwtToken(user string) (string, error) {
	if user == "" {
		return "", fmt.Errorf("GenerateJwtToken: user is empty")
	}
	if user == "admin" {
		return "", fmt.Errorf("GenerateJwtToken: user is admin")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user": user,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	// Подписываем токен с использованием приватного ключа
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return signedToken, nil
}

func GenerateJwtHandler(w http.ResponseWriter, r *http.Request) {
	token, err := GenerateJwtToken(r.URL.Query().Get("user"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: token,
	})
	w.Write([]byte(token))
}

func GenerateJwtTokenHack(user string) (string, error) {
	if user == "" {
		return "", fmt.Errorf("GenerateJwtToken: user is empty")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	log.Println(publicKeyBytes)
	signedToken, err := token.SignedString(publicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return signedToken, nil
}

func GenerateJwtHandlerHack(w http.ResponseWriter, r *http.Request) {
	token, err := GenerateJwtTokenHack(r.URL.Query().Get("user"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: token,
	})
	w.Write([]byte(token))
}

func VulnerableValidate(r *http.Request) (jwt.MapClaims, error) {
	tokenStr := GetJwtToken(r)
	if tokenStr == "" {
		return jwt.MapClaims{}, fmt.Errorf("GetJwtToken: no token found")
	}
	fmt.Println("Token:", tokenStr)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// для hmac нужны bytes
		return publicKeyBytes, nil // вот тут доверяем указанному в токене алгоритму и используем публичный RSA ключ как секрет для HMAC
	})
	if err != nil {
		if err.Error() == "key is of invalid type" {
			fmt.Println("try new key")
			token, err = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// для rsa нужен *rsa.PublicKey
				return publicKey, nil 
			})
			if err != nil {
				log.Printf("Ошибка при парсинге токена: %v %s\n", err, tokenStr)
				return jwt.MapClaims{}, fmt.Errorf("Ошибка при парсинге токена: %s", tokenStr)
			}
		} else {
			log.Printf("Ошибка при парсинге токена: %v %s\n", err, tokenStr)
			return jwt.MapClaims{}, fmt.Errorf("Ошибка при парсинге токена: %s", tokenStr)
		}
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if token.Valid && ok {
		return claims, nil
	}
	log.Printf("Ошибка при парсинге токена: %v %s\n", err, tokenStr)
	return jwt.MapClaims{}, fmt.Errorf("Ошибка при парсинге токена: %s", tokenStr)
}

func sendHttpFlag(w http.ResponseWriter, r *http.Request) {
	claims, err := VulnerableValidate(r)
	if err != nil {
		sendErr(&w, err)
		return
	}
	userString, ok := claims["user"].(string)
	if !ok {
		sendErr(&w, fmt.Errorf("field \"user\" not in jwt token"))
		return
	}
	if userString != "admin" {
		sendErr(&w, fmt.Errorf("Forbidden, only \"admin\" can view flag"))
		fmt.Println("VulnerableValidate - return", ok, err)
		return
	}
	t, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err := t.Execute(w, map[string]interface{}{
		"msg": os.Getenv("CTF_FLAG"),
	}); err != nil {
		log.Println(err)
		return
	}
}

func sendErr(w *http.ResponseWriter, errCustom error) {
	t, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err := t.Execute(*w, map[string]interface{}{
		"msg": errCustom.Error(),
	}); err != nil {
		log.Println(err)
		return
	}
}

func main() {


	mux := http.NewServeMux()

	mux.Handle("/static/server/", http.StripPrefix("/static/server/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", sendHttpFlag) // получить флаг

	// узнать публичный ключ
	mux.HandleFunc("/public-key", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(publicKeyBytes))
	})

	// получить токен
	mux.HandleFunc("/auth", GenerateJwtHandler)
	// тестовая ручка, сразу получает "хакнутый" токен, подписанный hmac
	mux.HandleFunc("/authHack", GenerateJwtHandlerHack)

	log.Println("Listening...")
	log.Println("public-key: ", string(publicKeyBytes))
	log.Fatal(http.ListenAndServe("0.0.0.0:8081", mux))
}
