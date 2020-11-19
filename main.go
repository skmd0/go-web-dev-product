package main

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go-web-dev/controllers"
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

	us := setupUserService(dsn)
	//us.AutoMigrate()
	//us.DestructiveReset()

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

func setupUserService(dsn string) user.UserService {
	us, err := user.NewUserService(dsn)
	if err != nil {
		panic(err)
	}
	return us
}

func setupUserController(us user.UserService) *controllers.Users {
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
