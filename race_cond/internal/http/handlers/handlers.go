package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
)


func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	_, err := db.Exec("INSERT INTO users (username) VALUES (?)", username)
	if err != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	fmt.Fprintf(w, "User %s registered", username)
}

func bonusHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	var gotBonus int
	err := db.QueryRow("SELECT got_bonus FROM users WHERE username = ?", username).Scan(&gotBonus)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if gotBonus == 1 {
		http.Error(w, "Bonus already claimed", http.StatusForbidden)
		return
	}

	_, err = db.Exec("UPDATE users SET balance = balance + 100, got_bonus = 1 WHERE username = ?", username)
	if err != nil {
		http.Error(w, "Failed to update balance", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Bonus granted to %s", username)
}

func flagHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	var balance int
	err := db.QueryRow("SELECT balance FROM users WHERE username = ?", username).Scan(&balance)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if balance < 200 {
		http.Error(w, "Not enough balance", http.StatusForbidden)
		return
	}

	_, err = db.Exec("UPDATE users SET balance = balance - 200 WHERE username = ?", username)
	if err != nil {
		http.Error(w, "Failed to purchase flag", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "FLAG{race_condition_ctf}")
}
