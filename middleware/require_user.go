package middleware

import (
	"fmt"
	"go-web-dev/context"
	"go-web-dev/models/user"
	"log"
	"net/http"
)

type RequireUser struct {
	user.UserService
}

func (mw *RequireUser) Apply(next http.Handler) http.Handler {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		usr, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		ctx := r.Context()
		ctx = context.WithUser(ctx, usr)
		r = r.WithContext(ctx)
		fmt.Println("User found:", usr)
		next(w, r)
	})
}
