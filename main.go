package main

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go-web-dev/controllers"
	"go-web-dev/models"
	"net/http"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "testtest"
	dbName   = "postgres"
)

func main() {
	var err error
	dsn := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		host, port, user, password, dbName)

	us := setupUserService(dsn)
	us.AutoMigrate()
	//us.DestructiveReset()
	//CreateAFakeUser(us)

	staticC := setupStaticController()
	usersC := setupUserController(us)

	r := mux.NewRouter()
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")
	err = http.ListenAndServe(":3000", r)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to start a HTTP server: %s", err.Error())
		panic(errMsg)
	}
}

func CreateAFakeUser(us *models.UserService) {
	must(us.DestructiveReset())
	user := &models.User{
		Name:     "Domen Skamlic",
		Email:    "domen@skamlic.com",
		Password: "testtest",
	}
	err := us.Create(user)
	if err != nil {
		panic(err)
	}
}

func setupUserService(dsn string) *models.UserService {
	us, err := models.NewUserService(dsn)
	if err != nil {
		panic(err)
	}
	return us
}

func setupUserController(us *models.UserService) *controllers.Users {
	usersC, err := controllers.NewUsers(us)
	if err != nil {
		panic(err)
	}
	return usersC
}

func setupStaticController() *controllers.StaticViews {
	staticC, err := controllers.NewStatic()
	if err != nil {
		panic(err)
	}
	return staticC
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
