package images

import (
	"accessCtf/internal/app"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func GetIndexPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/images/index.html")
	if err != nil {
		log.Println(err)
		return
	}

	if err := t.Execute(w, map[string]interface{}{"isLogined": false}); err != nil {
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
