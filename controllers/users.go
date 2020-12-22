package controllers

import (
	"go-web-dev/errs"
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
	u.NewView.Render(w, r, nil)
}

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var signUpForm UserSignUp
	err := parseForm(r, &signUpForm)
	if err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}
	usr := user.User{
		Name:     signUpForm.Name,
		Email:    signUpForm.Email,
		Password: signUpForm.Password,
	}
	err = u.us.Create(&usr)
	if err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}
	err = u.signIn(w, &usr)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	userDB, err := u.us.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case errs.ErrNotFound:
			vd.SetAlertErr("Invalid email address")
		case errs.ErrPasswordIncorrect:
			vd.SetAlertErr("Invalid password")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return
	}
	err = u.signIn(w, userDB)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
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
