package internal

import (
	"go-web-dev/models/user"
	"net/http"
	"strings"
)

type User struct {
	user.UserService
}

func (mw *User) Apply(next http.Handler) http.Handler {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// when we serve static assets we do not need to look up token
		// this is done to avoid unnecessary DB lookups when we are fetching bunch of images in gallery
		if path := r.URL.Path; strings.HasPrefix(path, "/assets/") || strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}
		usr, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}
		ctx := r.Context()
		ctx = WithUser(ctx, usr)
		r = r.WithContext(ctx)
		next(w, r)
	}
}

// RequireUser can only be used if User is run before it
type RequireUser struct {
	User
}

func (mw *RequireUser) Apply(next http.Handler) http.Handler {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr := GetUser(r.Context())
		if usr == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}
