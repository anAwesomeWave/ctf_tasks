package images

import (
	"accessCtf/internal/app"
	midauth "accessCtf/internal/http/middleware/auth"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
)

func GetIndexPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/images/index.html")
	if err != nil {
		log.Println(err)
		return
	}
	_, found := midauth.UserFromContext(r.Context())
	if err := t.Execute(w, map[string]interface{}{"isLogined": found}); err != nil {
		log.Println(err)
		return
	}
}

func GetUploadPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/images/upload.html")
	if err != nil {
		log.Println(err)
		return
	}

	if err := t.Execute(w, map[string]interface{}{"isLogined": false}); err != nil {
		log.Println(err)
		return
	}
}

func PostUploadImage(imageApp app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer func() { _ = file.Close() }()
		_ = handler
		if err := imageApp.SaveImage(file, "test-user", 1, app.DefaultImage); err != nil {
			fmt.Println(err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Success"))
	}
}

func GetImage(imageApp app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "userId")
		imageId := chi.URLParam(r, "imageId")
		uiImageId, err := strconv.ParseUint(imageId, 10, 32)
		if err != nil {
			http.Error(w, "Invalid image id", http.StatusBadRequest)
			return
		}
		file, err := imageApp.LoadImage(userId, uiImageId, app.DefaultImage)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal error", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		data, err := io.ReadAll(file)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}
}
