package controllers

import (
	"github.com/gorilla/schema"
	"net/http"
)

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	dec := schema.NewDecoder()
	// this is needed to ignore the CSRF hidden input element
	dec.IgnoreUnknownKeys(true)
	return dec.Decode(dst, r.PostForm)
}
