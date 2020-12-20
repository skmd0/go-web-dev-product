package controllers

import (
	"fmt"
	"github.com/gorilla/mux"
	"go-web-dev/context"
	"go-web-dev/errs"
	"go-web-dev/models/gallery"
	"go-web-dev/views"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	GalleryShowName = "show_gallery"
	GalleryEditName = "edit_gallery"

	maxMultipartMem = 1 << 20 // 1 megabyte
)

func NewGallery(gs gallery.GalleryService, r *mux.Router) (*Gallery, error) {
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
		r:         r,
	}, nil
}

type Gallery struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        gallery.GalleryService
	r         *mux.Router
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
		g.New.Render(w, r, vd)
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
		g.New.Render(w, r, vd)
		return
	}
	glrStrID := strconv.Itoa(int(glr.ID))
	url, err := g.r.Get(GalleryEditName).URL("id", glrStrID)
	if err != nil {
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
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	var form GalleryForm
	vd.Yield = glr
	err = parseForm(r, &form)
	if err != nil {
		log.Println(err)
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	glr.Title = form.Title
	err = g.gs.Update(glr)
	if err != nil {
		log.Println(err)
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
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	err = g.gs.Delete(glr.ID)
	if err != nil {
		log.Println(err)
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

	galleryPath := fmt.Sprintf("../images/galleries/%v/", glr.ID)
	err = os.MkdirAll(galleryPath, 0755)
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

		dst, err := os.Create(galleryPath + f.Filename)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}

		_, err = io.Copy(dst, file)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}

		_, _ = fmt.Fprintln(w, "files successfully uploaded")
		_ = dst.Close()
		_ = file.Close()
	}
}

func (g *Gallery) galleryByID(w http.ResponseWriter, r *http.Request) (*gallery.Gallery, error) {
	vars := mux.Vars(r)
	strID := vars["id"]
	id, err := strconv.Atoi(strID)
	if err != nil {
		http.Error(w, "Invalid gallery ID.", http.StatusNotFound)
		return nil, err
	}
	glr, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case errs.ErrNotFound:
			http.Error(w, "Gallery not found.", http.StatusNotFound)
		default:
			http.Error(w, "Oops! Something went wrong.", http.StatusInternalServerError)
		}
		return nil, err
	}
	return glr, nil
}
