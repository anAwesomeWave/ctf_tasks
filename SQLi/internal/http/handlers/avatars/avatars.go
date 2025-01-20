package avatars

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"sqli/internal/app"
	"sqli/internal/http/common"
	midauth "sqli/internal/http/middleware/auth"
	"sqli/internal/storage"
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
		r.Body = http.MaxBytesReader(w, r.Body, imageApp.GetMaxFileBytes())
		if err := r.ParseMultipartForm(imageApp.GetMaxFileBytes()); err != nil {
			common.ServeError(w, http.StatusBadRequest, "Bad Request. File too large.", ok)
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
		iAvatarId, err := strconv.ParseInt(avatarId, 10, 32)
		if err != nil {
			common.ServeError(
				w,
				http.StatusBadRequest,
				"Invalid avatar id",
				userFound,
			)
			return
		}
		file, err := imageApp.LoadImage(userId, iAvatarId, app.Avatar)
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
