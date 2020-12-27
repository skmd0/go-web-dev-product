package controllers

import (
	"fmt"
	"github.com/gorilla/mux"
	"go-web-dev/context"
	"go-web-dev/internal"
	"go-web-dev/models/gallery"
	"go-web-dev/models/images"
	"go-web-dev/views"
	"log"
	"net/http"
	"strconv"
)

const (
	GalleryShowName = "show_gallery"
	GalleryEditName = "edit_gallery"

	maxMultipartMem = 1 << 20 // 1 megabyte
)

func NewGallery(gs gallery.GalleryService, is images.ImageService, r *mux.Router) (*Gallery, error) {
	newGalleryView, err := views.NewView("bulma", "gallery/new")
	if err != nil {
		return nil, err
	}
	showView, err := views.NewView("bulma", "gallery/show")
	if err != nil {
		return nil, err
	}
	editView, err := views.NewView("bulma", "gallery/edit")
	if err != nil {
		return nil, err
	}
	indexView, err := views.NewView("bulma", "gallery/index")
	if err != nil {
		return nil, err
	}
	return &Gallery{
		New:       newGalleryView,
		ShowView:  showView,
		EditView:  editView,
		IndexView: indexView,
		gs:        gs,
		is:        is,
		r:         r,
	}, nil
}

type Gallery struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        gallery.GalleryService
	is        images.ImageService
	r         *mux.Router
}
type GalleryForm struct {
	Title string `schema:"title"`
}

// Create creates a new Gallery resource
func (g *Gallery) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var galleryForm GalleryForm
	err := parseForm(r, &galleryForm)
	if err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	user := context.User(r.Context())
	glr := gallery.Gallery{
		UserID: user.ID,
		Title:  galleryForm.Title,
	}
	err = g.gs.Create(&glr)
	if err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	glrStrID := strconv.Itoa(int(glr.ID))
	url, err := g.r.Get(GalleryEditName).URL("id", glrStrID)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// GET /galleries
func (g *Gallery) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	galleries, err := g.gs.ByUserID(user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yield = galleries
	g.IndexView.Render(w, r, vd)
}

// GET /gallery/:id
func (g *Gallery) Show(w http.ResponseWriter, r *http.Request) {
	glr, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = glr
	g.ShowView.Render(w, r, vd)
}

// GET /gallery/:id/edit
func (g *Gallery) Edit(w http.ResponseWriter, r *http.Request) {
	glr, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if glr.UserID != user.ID {
		log.Println(err)
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	vd.Yield = glr
	g.EditView.Render(w, r, vd)
}

// GET /gallery/:id/update
func (g *Gallery) Update(w http.ResponseWriter, r *http.Request) {
	glr, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if user == nil || glr.UserID != user.ID {
		log.Println(err)
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	var form GalleryForm
	vd.Yield = glr
	err = parseForm(r, &form)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	glr.Title = form.Title
	err = g.gs.Update(glr)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	url := fmt.Sprintf("/gallery/%v", glr.ID)
	http.Redirect(w, r, url, http.StatusFound)
}

// GET /gallery/:id/delete
func (g *Gallery) Delete(w http.ResponseWriter, r *http.Request) {
	glr, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if user == nil || glr.UserID != user.ID {
		log.Println(err)
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	err = g.gs.Delete(glr.ID)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// POST /gallery/:id/images
func (g *Gallery) ImageUpload(w http.ResponseWriter, r *http.Request) {
	glr, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if user == nil || glr.UserID != user.ID {
		log.Println(err)
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	var vd views.Data
	vd.Yield = glr
	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		err = g.is.Create(glr.ID, file, f.Filename)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		_ = file.Close()
	}
	url, err := g.r.Get(GalleryEditName).URL("id", fmt.Sprintf("%v", glr.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (g *Gallery) galleryByID(w http.ResponseWriter, r *http.Request) (*gallery.Gallery, error) {
	vars := mux.Vars(r)
	strID := vars["id"]
	id, err := strconv.Atoi(strID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid gallery ID.", http.StatusNotFound)
		return nil, err
	}
	glr, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case internal.ErrNotFound:
			log.Println(err)
			http.Error(w, "Gallery not found.", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Oops! Something went wrong.", http.StatusInternalServerError)
		}
		return nil, err
	}
	imgs, err := g.is.ByGalleryID(glr.ID)
	glr.Images = imgs
	return glr, nil
}

// POST /gallery/:id/images/:filename/delete
func (g *Gallery) ImageDelete(w http.ResponseWriter, r *http.Request) {
	glr, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if user == nil || glr.UserID != user.ID {
		log.Println(err)
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	filename := mux.Vars(r)["filename"]
	i := &images.Image{
		GalleryID: glr.ID,
		Filename:  filename,
	}
	err = g.is.Delete(i)
	if err != nil {
		var vd views.Data
		vd.Yield = glr
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	glrStrID := strconv.Itoa(int(glr.ID))
	url, err := g.r.Get(GalleryEditName).URL("id", glrStrID)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}
