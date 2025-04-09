package handlers

import (
	"html/template"
	"log"
	"net/http"
	"race_cond/internal/http/common"
	midauth "race_cond/internal/http/middleware"
	"race_cond/internal/storage"
)

func GetIndexPage(w http.ResponseWriter, r *http.Request) {
	user, isLogined := midauth.UserFromContext(r.Context())
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/common/index.html")
	if err != nil {
		log.Println(err)
		return
	}
	balance := 0
	if user != nil {
		balance = user.Balance
	}
	if err := t.Execute(w, map[string]interface{}{"isLogined": isLogined, "balance": balance}); err != nil {
		log.Println(err)
		return
	}
}

func GetBonusPage(w http.ResponseWriter, r *http.Request) {
	_, isLogined := midauth.UserFromContext(r.Context())
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/common/bonus.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err := t.Execute(w, map[string]interface{}{"isLogined": isLogined}); err != nil {
		log.Println(err)
		return
	}
}

func GetBonus(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, isLogined := midauth.UserFromContext(r.Context())
		if user.GotBonus != 0 {
			common.ServeError(
				w,
				http.StatusForbidden,
				"You have already got the bonus!!!!!! stop wasting server's resources!",
				isLogined,
			)
			return
		}
		user, err := strg.UpdateBalance(user.Id)
		if err != nil {
			log.Println(err)
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"InternalError",
				isLogined,
			)
		}
		http.Redirect(
			w,
			r,
			"/",
			http.StatusSeeOther,
		)
	}
}

//func bonusHandler(w http.ResponseWriter, r *http.Request) {
//	username := r.URL.Query().Get("username")
//	if username == "" {
//		http.Error(w, "Username required", http.StatusBadRequest)
//		return
//	}
//
//	var gotBonus int
//	err := db.QueryRow("SELECT got_bonus FROM users WHERE username = ?", username).Scan(&gotBonus)
//	if err != nil {
//		http.Error(w, "User not found", http.StatusNotFound)
//		return
//	}
//
//	if gotBonus == 1 {
//		http.Error(w, "Bonus already claimed", http.StatusForbidden)
//		return
//	}
//
//	_, err = db.Exec("UPDATE users SET balance = balance + 100, got_bonus = 1 WHERE username = ?", username)
//	if err != nil {
//		http.Error(w, "Failed to update balance", http.StatusInternalServerError)
//		return
//	}
//
//	fmt.Fprintf(w, "Bonus granted to %s", username)
//}
//
//func flagHandler(w http.ResponseWriter, r *http.Request) {
//	username := r.URL.Query().Get("username")
//	if username == "" {
//		http.Error(w, "Username required", http.StatusBadRequest)
//		return
//	}
//
//	var balance int
//	err := db.QueryRow("SELECT balance FROM users WHERE username = ?", username).Scan(&balance)
//	if err != nil {
//		http.Error(w, "User not found", http.StatusNotFound)
//		return
//	}
//
//	if balance < 200 {
//		http.Error(w, "Not enough balance", http.StatusForbidden)
//		return
//	}
//
//	_, err = db.Exec("UPDATE users SET balance = balance - 200 WHERE username = ?", username)
//	if err != nil {
//		http.Error(w, "Failed to purchase flag", http.StatusInternalServerError)
//		return
//	}
//
//	fmt.Fprintf(w, "FLAG{race_condition_ctf}")
//}
