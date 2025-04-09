package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"net/http"
	"race_cond/internal/http/common"
	"race_cond/internal/http/handlers"
	midauth "race_cond/internal/http/middleware"
	"race_cond/internal/storage"
)

func setUpRouter(strg storage.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // не падать при панике
	router.Use(middleware.URLFormat) // удобно брать из урлов данные
	router.Use(middleware.StripSlashes)

	router.Handle("/static/server/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.Get("/users/logout", handlers.Logout) // unprotected
	router.Group(func(authR chi.Router) {
		authR.Use(jwtauth.Verifier(handlers.TokenAuth))
		authR.Use(midauth.GetUserByJwtToken(strg))

		authR.NotFound(func(w http.ResponseWriter, r *http.Request) {
			common.ServeError(w, http.StatusNotFound, "Not Found", false)
		})

		authR.Get("/", handlers.GetIndexPage)

		authR.Get("/users/signup", handlers.GetSignUpPage)
		authR.Post("/users/signup", handlers.PostSignUpPage(strg))
		authR.Get("/users/login", handlers.GetLoginPage)
		authR.Post("/users/login", handlers.PostLoginPage(strg))

		authR.Group(func(bonusR chi.Router) {
			bonusR.Use(midauth.CustomAuthenticator(handlers.TokenAuth))
			bonusR.Get("/bonus", handlers.GetBonusPage)
			bonusR.Post("/bonus", handlers.GetBonus(strg))
		})
	})
	return router
}
