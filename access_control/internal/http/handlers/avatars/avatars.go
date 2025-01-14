package avatars

import (
	"accessCtf/internal/app"
	"accessCtf/internal/http/common"
	midauth "accessCtf/internal/http/middleware/auth"
	"accessCtf/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strconv"
)

func PostUploadAvatar(imageApp app.App, strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := midauth.UserFromContext(r.Context())
		if !ok || user == nil {
			common.ServeError(w, 401, "Unauthorized!", false)
			return
		}
		maxAvataridForUser, err := strg.GetMaxUserAvatarId(user.Id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				maxAvataridForUser = 0
			} else {
				log.Println(err)
				common.ServeError(w, 500, "Internal error", user != nil)
				return
			}
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("myFile")
		if err != nil {
			log.Println("Error Retrieving the File")
			log.Println(err)
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"Error uploading image.",
				false,
			)
			return
		}
		defer func() { _ = file.Close() }()
		_ = handler
		path, err := imageApp.SaveImage(file, user.Id.String(), maxAvataridForUser+1, app.Avatar)
		if err != nil {
			log.Println(err)
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"Error saving image.",
				false,
			)
			return
		}

		// store to db
		_, err = strg.InsertAvatar(user.Id, maxAvataridForUser+1, path)
		if err != nil {
			log.Println(err)
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"Error uploading image.",
				false,
			)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func GetAvatar(imageApp app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, userFound := midauth.UserFromContext(r.Context())
		userId := chi.URLParam(r, "userId")
		avatarId := chi.URLParam(r, "avatarId")
		uiAvatarId, err := strconv.ParseUint(avatarId, 10, 32)
		if err != nil {
			common.ServeError(
				w,
				http.StatusBadRequest,
				"Invalid avatar id",
				userFound,
			)
			return
		}
		file, err := imageApp.LoadImage(userId, uiAvatarId, app.Avatar)
		if err != nil {
			log.Println(err)
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"Internal error",
				userFound,
			)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		data, err := io.ReadAll(file)
		if err != nil {
			log.Println(err)
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"Internal error",
				userFound,
			)
			return
		}
		w.Write(data)
	}
}
