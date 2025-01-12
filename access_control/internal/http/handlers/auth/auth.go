package auth

import (
	"accessCtf/internal/http/common"
	"accessCtf/internal/storage"
	"errors"
	"html/template"
	"log"
	"net/http"
)

func GetSignUpPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/auth/signup.html")
	if err != nil {
		log.Println(err)
		return
	}

	if err := t.Execute(w, map[string]interface{}{"isLogined": false}); err != nil {
		log.Println(err)
		return
	}
}

func PostSignUpPage(strg storage.Storage) http.HandlerFunc {
	const fn = "storage.CreateUser"

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"Unable to parse form.",
				false,
			)
			return
		}
		if _, ok := r.Form["login"]; !ok {
			common.ServeError(
				w,
				http.StatusBadRequest,
				"Bad form. Unable to find `login` field.",
				false,
			)
			return
		}
		if _, ok := r.Form["password"]; !ok {
			common.ServeError(
				w,
				http.StatusBadRequest,
				"Bad form. Unable to find `password` field.",
				false,
			)
			return
		}
		login, password := r.Form["login"][0], r.Form["password"][0]
		if len(login) == 0 || len(password) == 0 {
			log.Println(login, password)
			common.ServeError(
				w,
				http.StatusBadRequest,
				"user Parsing error. fields are empty.",
				false,
			)
			return
		}
		if _, err := strg.CreateUser(login, password); err != nil {
			if errors.Is(err, storage.ErrExists) {
				common.ServeError(
					w,
					http.StatusBadRequest,
					"User already exists.",
					false,
				)
				return
			}
			log.Printf("%s:%v\n", fn, err)
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"Failed to create user.",
				false,
			)
			return
		}
		http.Redirect(
			w,
			r,
			"/users/login",
			http.StatusSeeOther,
		)
	}
}

func GetLoginPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/auth/login.html")
	if err != nil {
		log.Println(err)
		return
	}

	if err := t.Execute(w, map[string]interface{}{"isLogined": false}); err != nil {
		log.Println(err)
		return
	}
}
