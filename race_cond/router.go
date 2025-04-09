package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"net/http"
	
)

func setUpRouter(imagesApp app.App, strg storage.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // не падать при панике
	router.Use(middleware.URLFormat) // удобно брать из урлов данные
	router.Use(middleware.StripSlashes)

	router.Handle("/static/server/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.Get("/users/logout", auth.Logout) // unprotected
	router.Group(func(authR chi.Router) {
		authR.Use(jwtauth.Verifier(auth.TokenAuth))
		authR.Use(midauth.GetUserByJwtToken(strg))

		authR.NotFound(func(w http.ResponseWriter, r *http.Request) {
			common.ServeError(w, http.StatusNotFound, "Not Found", false)
		})

		authR.Get("/", images.GetIndexPage(strg))

		authR.Get("/users/signup", auth.GetSignUpPage)
		authR.Post("/users/signup", auth.PostSignUpPage(strg))
		authR.Get("/users/login", auth.GetLoginPage)
		authR.Post("/users/login", auth.PostLoginPage(strg))

		authR.Route("/static/images", func(r chi.Router) {
			r.Get("/{userId}/{imageId}", images.GetImage(strg, imagesApp))
			r.Route("/upload", func(subR chi.Router) {
				subR.Use(midauth.CustomAuthenticator(auth.TokenAuth))
				subR.Get("/", images.GetUploadPage("images", "Image"))
				subR.Post("/", images.PostUploadImage(imagesApp, strg))
			})
		})
		authR.Route("/static/avatars", func(r chi.Router) {
			r.Get("/{userId}/{avatarId}", avatars.GetAvatar(imagesApp))
			r.Route("/upload", func(subR chi.Router) {
				subR.Use(midauth.CustomAuthenticator(auth.TokenAuth))
				subR.Get("/", images.GetUploadPage("avatars", "Avatar"))
				subR.Post("/", avatars.PostUploadAvatar(imagesApp, strg))
			})
		})
		authR.Group(func(usersR chi.Router) {
			usersR.Use(midauth.CustomAuthenticator(auth.TokenAuth))
			authR.Get("/users/me", users.GetMePage(strg))
			authR.Post("/users/me/edit", users.PostUpdateAccount(imagesApp, strg))
		})
	})
	return router
}
