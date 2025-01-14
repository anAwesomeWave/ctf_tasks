package users

import (
	"accessCtf/internal/http/common"
	midauth "accessCtf/internal/http/middleware/auth"
	"accessCtf/internal/storage"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func GetMePage(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, isLogined := midauth.UserFromContext(r.Context())
		if !isLogined || user == nil {
			common.ServeError(w, http.StatusUnauthorized, "Unauthorized!", isLogined)
			return
		}
		isAvatarExist := true
		avatarPath, err := strg.GetLastUploadAvatar(user.Id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				isAvatarExist = false
			} else {
				log.Println(err)
				common.ServeError(w, 500, "Internal error", user != nil)
				return
			}
		}
		if isAvatarExist {
			afterAvatar, found := strings.CutPrefix(avatarPath, "./static/users/upload/")
			if !found {
				log.Println("ERROR. STRANGE STRING PATTERN:", avatarPath)
				common.ServeError(w, http.StatusInternalServerError, "Internal server error", isLogined)
			}
			avatarPath = "/static/" + afterAvatar
		}
		images, err := strg.GetAllUserImages(user.Id)
		if err != nil {
			log.Println(err)
			common.ServeError(w, http.StatusInternalServerError, "Internal server error", isLogined)
			return
		}
		for ind := range images {
			after, found := strings.CutPrefix(images[ind].Path, "./static/users/upload/")
			if !found {
				log.Println("ERROR. STRANGE STRING PATTERN:", images[ind].Path)
				common.ServeError(w, http.StatusInternalServerError, "Internal server error", isLogined)
			}
			images[ind].Path = "/static/images/" + after
		}
		t, err := template.ParseFiles("./templates/common/base.html", "./templates/users/me.html")
		if err != nil {
			log.Println(err)
			return
		}
		if err := t.Execute(w, map[string]interface{}{
			"isLogined":     isLogined,
			"user":          user,
			"images":        images,
			"avatarsPath":   avatarPath,
			"isAvatarExist": isAvatarExist,
		}); err != nil {
			log.Println(err)
			return
		}
	}
}
