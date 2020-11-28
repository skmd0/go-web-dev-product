package middleware

import (
	"go-web-dev/context"
	"go-web-dev/models/user"
	"net/http"
)

type User struct {
	user.UserService
}

func (mw *User) Apply(next http.Handler) http.Handler {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		ctx = context.WithUser(ctx, usr)
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
		usr := context.User(r.Context())
		if usr == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}
