package controllers

import (
	"fmt"
	"go-web-dev/models/user"
	"go-web-dev/rand"
	"go-web-dev/views"
	"log"
	"net/http"
)

func NewUsers(us user.UserService) (*Users, error) {
	signUpView, err := views.NewView("bulma", "users/new")
	if err != nil {
		return nil, err
	}
	loginView, err := views.NewView("bulma", "users/login")
	if err != nil {
		return nil, err
	}
	return &Users{
		NewView:   signUpView,
		LoginView: loginView,
		us:        us,
	}, nil

}

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        user.UserService
}

type UserSignUp struct {
	Name           string `schema:"name"`
	Email          string `schema:"email"`
	Password       string `schema:"password"`
	RepeatPassword string `schema:"repeat-password"`
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	err := u.NewView.Render(w, nil)
	if err != nil {
		fmt.Println("ERR: failed to render new.gohtml")
	}
}

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var signUpForm UserSignUp
	err := parseForm(r, &signUpForm)
	if err != nil {
		log.Println(err)
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: views.AlertMsgGeneric,
		}
		u.NewView.Render(w, vd)
		return
	}
	usr := user.User{
		Name:     signUpForm.Name,
		Email:    signUpForm.Email,
		Password: signUpForm.Password,
	}
	err = u.us.Create(&usr)
	if err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: err.Error(),
		}
		u.NewView.Render(w, vd)
		return
	}
	err = u.signIn(w, &usr)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userDB, err := u.us.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			fmt.Fprintln(w, "Invalid email address")
		case user.ErrPasswordIncorrect:
			fmt.Fprintln(w, "Invalid password")
		default:
			http.Error(w, "Something went wrong. Please try again later.", http.StatusInternalServerError)
		}
		return
	}
	err = u.signIn(w, userDB)
	if err != nil {
		http.Error(w, "Something went wrong. Please try again later.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)
}

func (u *Users) signIn(w http.ResponseWriter, user *user.User) error {
	if user.Remember == "" {
		token, err := rand.GenerateRememberToken(rand.RememberTokenBytes)
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to find the user: %s", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, user)
}
