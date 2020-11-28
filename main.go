package main

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go-web-dev/controllers"
	"go-web-dev/middleware"
	"go-web-dev/models"
	"go-web-dev/models/gallery"
	"go-web-dev/models/user"
	"net/http"
)

const (
	host     = "localhost"
	port     = 5432
	pq_user  = "postgres"
	password = "testtest"
	dbName   = "postgres"
)

func main() {
	var err error
	dsn := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		host, port, pq_user, password, dbName)
	services, err := models.NewServices(dsn)
	if err != nil {
		panic(err)
	}

	//services.DestructiveReset()
	err = services.AutoMigrate()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	staticC := setupStaticController()
	usersC := setupUserController(services.User)
	galleryC := setupGalleryController(services.Gallery, r)
	requireUserMw := middleware.RequireUser{UserService: services.User}

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")

	// Gallery routes
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleryC.Index)).Methods("GET")
	r.Handle("/gallery/new", requireUserMw.Apply(galleryC.New)).Methods("GET")
	r.HandleFunc("/gallery", requireUserMw.ApplyFn(galleryC.Create)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleryC.Edit)).Methods("GET").
		Name(controllers.GalleryEditName)
	r.HandleFunc("/gallery/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleryC.Update)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}/delete", requireUserMw.ApplyFn(galleryC.Delete)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}", galleryC.Show).Methods("GET").
		Name(controllers.GalleryShowName)

	err = http.ListenAndServe(":3000", r)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to start a HTTP server: %s", err.Error())
		panic(errMsg)
	}
}

func setupUserController(us user.UserService) *controllers.Users {
	usersC, err := controllers.NewUsers(us)
	if err != nil {
		panic(err)
	}
	return usersC
}

func setupGalleryController(gs gallery.GalleryService, r *mux.Router) *controllers.Gallery {
	galleryC, err := controllers.NewGallery(gs, r)
	if err != nil {
		panic(err)
	}
	return galleryC
}

func setupStaticController() *controllers.StaticViews {
	staticC, err := controllers.NewStatic()
	if err != nil {
		panic(err)
	}
	return staticC
}
