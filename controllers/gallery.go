package controllers

import (
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
	return &Gallery{
		New: newGalleryView,
		gs:  gs,
	}, nil

}

type Gallery struct {
	New *views.View
	gs  gallery.GalleryService
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
	glr := gallery.Gallery{Title: galleryForm.Title}
	err = g.gs.Create(&glr)
	if err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}
	http.Redirect(w, r, "/gallery/:id", http.StatusFound)
}
