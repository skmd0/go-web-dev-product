package controllers

import (
	"go-web-dev/context"
	"go-web-dev/models/gallery"
	"go-web-dev/views"
	"log"
	"net/http"
)

func NewGallery(gs gallery.GalleryService) (*Gallery, error) {
	newGalleryView, err := views.NewView("bulma", "gallery/new")
	if err != nil {
		return nil, err
	}
	showView, err := views.NewView("bulma", "gallery/show")
	if err != nil {
		return nil, err
	}
	return &Gallery{
		New:      newGalleryView,
		ShowView: showView,
		gs:       gs,
		r:        r,
	}, nil

}

type Gallery struct {
	New      *views.View
	ShowView *views.View
	gs       gallery.GalleryService
	r        *mux.Router
}
type GalleryForm struct {
	Title string `schema:"title"`
}

func (g *Gallery) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var galleryForm GalleryForm
	err := parseForm(r, &galleryForm)
	if err != nil {
		log.Println(err)
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}
	user := context.User(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	glr := gallery.Gallery{
		UserID: user.ID,
		Title:  galleryForm.Title,
	}
	err = g.gs.Create(&glr)
	if err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}
	http.Redirect(w, r, "/gallery/:id", http.StatusFound)
}
