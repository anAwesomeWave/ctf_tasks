package images

import (
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
