package auth

import (
	"accessCtf/internal/http/common"
	"accessCtf/internal/storage"
	"accessCtf/internal/storage/models"
	"context"
	"errors"
	"github.com/go-chi/jwtauth"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
	"log"
	"net/http"
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
				// за это отвечает jwtauth.Authenticator
				//http.Error(w, "Unauthorized", http.StatusUnauthorized)
				next.ServeHTTP(w, r)
				return
			}
			userIdString, ok := claims["user_id"].(string)
			if !ok {
				log.Printf("%v: Cannot get userId from claims user_id - %v", fn, claims["user_id"])

				common.ServeError(w, http.StatusInternalServerError, "Invalid token", true)
				return
			}
			userUUID, err := uuid.Parse(userIdString)
			if err != nil {
				log.Printf("%s: Cannot parse token string into UUID - %s", fn, userIdString)
				common.ServeError(w, http.StatusInternalServerError, "Invalid token", true)
				return
			}
			user, err := strg.GetUserById(userUUID)
			if err != nil {
				user = nil
				// за это отвечает jwtauth.Authenticator
				log.Printf("%s: User with id %v not found in database: %v\n", fn, userUUID, err)
				common.ServeError(w, http.StatusInternalServerError, "User not found in database", true)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (*models.Users, bool) {
	user, ok := ctx.Value(userContextKey).(*models.Users)
	return user, ok
}

func CustomAuthenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			if err != nil || token == nil || jwt.Validate(token) != nil {
				common.ServeError(w, 401, "Unauthorized! please, login to your account", false)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
