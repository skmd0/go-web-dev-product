package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go-web-dev/controllers"
	"go-web-dev/internal"
	"go-web-dev/middleware"
	"go-web-dev/models"
	"net/http"
	"os"
)

func main() {
	if err := run(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "%s\n", err)
		panic(err)
	}
}

func run() error {
	var err error

	configReq := flag.Bool("prod", false,
		"Provide this flag in production. This ensures that .config is provided before application starts.")
	flag.Parse()

	cfg := internal.LoadConfig(*configReq)
	dbCfg := cfg.Database
	services, err := models.NewServices(
		models.WithGorm(dbCfg.ConnectionInfo()),
		models.WithUser(cfg.HMACKey, cfg.Pepper),
		models.WithGallery(),
		models.WithImages(),
	)
	if err != nil {
		return err
	}

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

	galleryC, err := controllers.NewGallery(services.Gallery, services.Images, r)
	if err != nil {
		return err
	}

	csrfToken, err := internal.GenerateRememberToken(32)
	if err != nil {
		return err
	}
	csrfMw := csrf.Protect([]byte(csrfToken), csrf.Secure(cfg.IsProd()))
	userMw := middleware.User{ServiceUser: services.User}
	requireUserMw := middleware.RequireUser{User: userMw}

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/logout", requireUserMw.ApplyFn(usersC.Logout)).Methods("POST")
	r.Handle("/forgot", usersC.ForgotPwView).Methods("GET")
	r.HandleFunc("/forgot", usersC.InitiateReset).Methods("POST")
	r.HandleFunc("/reset", usersC.ResetPw).Methods("GET")
	r.HandleFunc("/reset", usersC.CompleteReset).Methods("POST")

	// Gallery routes
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleryC.Index)).Methods("GET")
	r.Handle("/gallery/new", requireUserMw.Apply(galleryC.New)).Methods("GET")
	r.HandleFunc("/gallery", requireUserMw.ApplyFn(galleryC.Create)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleryC.Edit)).Methods("GET").
		Name(controllers.GalleryEditName)
	r.HandleFunc("/gallery/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleryC.Update)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}/delete", requireUserMw.ApplyFn(galleryC.Delete)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}", galleryC.Show).Methods("GET").Name(controllers.GalleryShowName)
	r.HandleFunc("/gallery/{id:[0-9]+}/images", requireUserMw.ApplyFn(galleryC.ImageUpload)).Methods("POST")
	r.HandleFunc("/gallery/{id:[0-9]+}/images/{filename}/delete", requireUserMw.ApplyFn(galleryC.ImageDelete)).
		Methods("POST")

	// image routes
	imageHandler := http.FileServer(http.Dir("../images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// assets
	assetHandler := http.StripPrefix("/assets/", http.FileServer(http.Dir("../assets/")))
	r.PathPrefix("/assets/").Handler(assetHandler)

	fmt.Printf("Starting the server on :%d\n", cfg.Port)
	address := fmt.Sprintf(":%d", cfg.Port)
	err = http.ListenAndServe(address, csrfMw(userMw.Apply(r)))
	if err != nil {
		return err
	}
	return nil
}
