package images

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"html/template"
	"io"
	"log"
	"net/http"
	"sqli/internal/app"
	"sqli/internal/http/common"
	midauth "sqli/internal/http/middleware/auth"
	"sqli/internal/storage"
	"strconv"
	"strings"
)

func GetIndexPage(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, isLogined := midauth.UserFromContext(r.Context())
		images, err := strg.GetAllImagesWithUserInfo()
		if err != nil {
			log.Println(err)
			common.ServeError(w, http.StatusInternalServerError, "Internal server error", isLogined)
			return
		}
		for ind := range images {
			after, found := strings.CutPrefix(images[ind].ImagePath, "./static/users/upload/")
			if !found {
				log.Println("ERROR. STRANGE STRING PATTERN:", images[ind].ImagePath)
				common.ServeError(w, http.StatusInternalServerError, "Internal server error", isLogined)
			}
			images[ind].ImagePath = "/static/images/" + after
			if images[ind].IsAdmin && !(isLogined && user.IsAdmin) {
				images[ind].ImagePath = "/static/images/default/1.jpeg"
			}
			if images[ind].AvatarPath != "" {
				afterAvatar, found := strings.CutPrefix(images[ind].AvatarPath, "./static/users/upload/")
				if !found {
					log.Println("ERROR. STRANGE STRING PATTERN:", images[ind].AvatarPath)
					common.ServeError(w, http.StatusInternalServerError, "Internal server error", isLogined)
				}
				images[ind].AvatarPath = "/static/" + afterAvatar
			}
		}
		t, err := template.ParseFiles("./templates/common/base.html", "./templates/images/index.html")
		if err != nil {
			log.Println(err)
			return
		}
		if err := t.Execute(w, map[string]interface{}{"isLogined": isLogined, "images": images}); err != nil {
			log.Println(err)
			return
		}
	}
}

func GetUploadPage(urlParam, filetype string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, isLogined := midauth.UserFromContext(r.Context())
		t, err := template.ParseFiles("./templates/common/base.html", "./templates/images/upload.html")
		if err != nil {
			log.Println(err)
			return
		}

		if err := t.Execute(w, map[string]interface{}{"isLogined": isLogined, "urlParam": urlParam, "filetype": filetype}); err != nil {
			log.Println(err)
			return
		}
	}
}
func PostUploadImage(imageApp app.App, strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
			1. получить юзера
			2. получить максимальный для него id картинки
			3. загрузить картинку, убедиться, что все ок
			4. обновить таблицу с картинками

			! возможно также проверить, сколько занимает его директория !
		*/
		user, ok := midauth.UserFromContext(r.Context())
		if !ok || user == nil {
			common.ServeError(w, 401, "Unauthorized!", false)
			return
		}
		maxImageidForUser, err := strg.GetMaxUserImageId(user.Id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				maxImageidForUser = 0
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
		path, err := imageApp.SaveImage(file, user.Id.String(), maxImageidForUser+1, app.DefaultImage)
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
		_, err = strg.InsertImage(user.Id, maxImageidForUser+1, path)
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

func GetImage(strg storage.Storage, imageApp app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, userFound := midauth.UserFromContext(r.Context())
		userId := chi.URLParam(r, "userId")
		imageId := chi.URLParam(r, "imageId")
		iImageId, err := strconv.ParseInt(imageId, 10, 32)
		if err != nil {
			common.ServeError(
				w,
				http.StatusBadRequest,
				"Invalid image id",
				user != nil,
			)
			return
		}
		isPublic := true
		if userId != "default" {
			userUUID, err := uuid.Parse(userId)
			if err != nil {
				log.Println(err)
				common.ServeError(
					w,
					http.StatusInternalServerError,
					"Internal error",
					user != nil,
				)
				return
			}
			isPublic, err = strg.IsImagePublic(userUUID)
			if err != nil {
				log.Println(err)
				common.ServeError(
					w,
					http.StatusInternalServerError,
					"Internal error",
					user != nil,
				)
				return
			}
		}
		if !(isPublic || (userFound && (user.IsAdmin || user.Id.String() == userId))) {
			common.ServeError(
				w,
				http.StatusUnauthorized,
				"you have no rights to view this image!!!",
				user != nil,
			)
			return
		}
		file, err := imageApp.LoadImage(userId, iImageId, app.DefaultImage)
		if err != nil {
			log.Println(err)
			common.ServeError(
				w,
				http.StatusInternalServerError,
				"Internal error",
				user != nil,
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
				user != nil,
			)
			return
		}
		w.Write(data)
	}
}
