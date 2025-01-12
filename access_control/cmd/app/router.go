package main

import (
	"accessCtf/internal/app"
	"accessCtf/internal/http/common"
	"accessCtf/internal/http/handlers/auth"
	"accessCtf/internal/http/handlers/images"
	midauth "accessCtf/internal/http/middleware/auth"
	"accessCtf/internal/storage"
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

	router.Use(jwtauth.Verifier(auth.TokenAuth))
	router.Use(midauth.GetUserByJwtToken(strg))

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		common.ServeError(w, http.StatusNotFound, "Not Found", false)
	})

	router.Handle("/static/server/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.Get("/static/{userId}/{imageId}", images.GetImage(imagesApp))

	router.Get("/", images.GetIndexPage)
	router.Get("/upload", images.GetUploadPage)
	router.Post("/upload", images.PostUploadImage(imagesApp))

	router.Get("/users/signup", auth.GetSignUpPage)
	router.Post("/users/signup", auth.PostSignUpPage(strg))
	router.Get("/users/login", auth.GetLoginPage)
	router.Post("/users/login", auth.PostLoginPage(strg))
	return router
}
