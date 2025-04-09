package auth

import (
	"context"
	"errors"
	"github.com/go-chi/jwtauth"
	"log"
	"net/http"
	"race_cond/internal/http/common"
	"race_cond/internal/storage"
	"race_cond/internal/storage/models"
	"strconv"
)

type contextKey string

const userContextKey = contextKey("user")

var UnauthorizedErr = errors.New("Unauthorized user")

func GetUserByJwtToken(strg storage.Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		const fn = "Middleware.Auth.GetUserByJwtToken"
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			userIdString, ok := claims["user_id"].(string)
			if !ok {
				log.Printf("%v: Cannot get userId from claims user_id - %v", fn, claims["user_id"])

				common.ServeError(w, http.StatusInternalServerError, "Invalid token", true)
				return
			}
			userId, err := strconv.ParseInt(userIdString, 10, 64)
			if err != nil {
				log.Printf("%s: Cannot parse token string into int64 - %s", fn, userIdString)
				common.ServeError(w, http.StatusInternalServerError, "Invalid token", true)
				return
			}
			user, err := strg.GetUserById(userId)
			if err != nil {
				user = nil
				log.Printf("%s: User with id %v not found in database: %v\n", fn, userId, err)
				common.ServeError(w, http.StatusInternalServerError, "User not found in database", true)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}
