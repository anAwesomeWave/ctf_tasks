package main

import (
	"accessCtf/internal/app"
	"accessCtf/internal/http/handlers/images"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func setUpRouter(imagesApp app.App /*db, app*/) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // не падать при панике
	router.Use(middleware.URLFormat) // удобно брать из урлов данные
	router.Use(middleware.StripSlashes)

	router.Handle("/static/server/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.Get("/static/{userId}/{imageId}", images.GetImage(imagesApp))

	router.Get("/", images.GetIndexPage)
	router.Get("/upload", images.GetUploadPage)
	router.Post("/upload", images.PostUploadImage(imagesApp))
	return router
}
