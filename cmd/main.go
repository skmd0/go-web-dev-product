package main

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go-web-dev/controllers"
	"go-web-dev/middleware"
	"go-web-dev/models"
	"net/http"
	"os"
)

const (
	host     = "localhost"
	port     = 5432
	pq_user  = "postgres"
	password = "testtest"
	dbName   = "postgres"
)

func main() {
	if err := run(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "%s\n", err)
		panic(err)
	}
}

func run() error {
	var err error
	dsn := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		host, port, pq_user, password, dbName)
	services, err := models.NewServices(dsn)
	if err != nil {
		return err
	}

	//services.DestructiveReset()
	err = services.AutoMigrate()
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	staticC, err := controllers.NewStatic()
	if err != nil {
		return err
	}

	usersC, err := controllers.NewUsers(services.User)
	if err != nil {
		return err
	}

	galleryC, err := controllers.NewGallery(services.Gallery, r)
	if err != nil {
		return err
	}

	userMw := middleware.User{UserService: services.User}
	requireUserMw := middleware.RequireUser{User: userMw}

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
	r.HandleFunc("/gallery/{id:[0-9]+}/images", requireUserMw.ApplyFn(galleryC.ImageUpload)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}/delete", requireUserMw.ApplyFn(galleryC.Delete)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}", galleryC.Show).Methods("GET").
		Name(controllers.GalleryShowName)

	err = http.ListenAndServe(":3000", userMw.Apply(r))
	if err != nil {
		return err
	}
	return nil
}
