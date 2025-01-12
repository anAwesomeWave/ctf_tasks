package common

import (
	"html/template"
	"log"
	"net/http"
)

func ServeError(w http.ResponseWriter, httpStatusCode int, message string, isLogined bool) {
	w.WriteHeader(httpStatusCode)
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/common/error.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err := t.Execute(w, map[string]interface{}{
		"isLogined":    isLogined,
		"errorCode":    httpStatusCode,
		"errorMessage": message,
	}); err != nil {
		http.Error(w, "Error templating", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	return
}
