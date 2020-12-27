package controllers

import (
	"fmt"
	"go-web-dev/context"
	"go-web-dev/internal"
	"go-web-dev/models/user"
	"go-web-dev/views"
	"log"
	"net/http"
	"time"
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
	forgotPwView, err := views.NewView("bulma", "users/forgot_pw")
	if err != nil {
		return nil, err
	}
	resetPwView, err := views.NewView("bulma", "users/reset_pw")
	if err != nil {
		return nil, err
	}
	return &Users{
		NewView:      signUpView,
		LoginView:    loginView,
		ForgotPwView: forgotPwView,
		ResetPwView:  resetPwView,
		us:           us,
	}, nil
}

type Users struct {
	NewView      *views.View
	LoginView    *views.View
	ForgotPwView *views.View
	ResetPwView  *views.View
	us           user.UserService
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
	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Welcome to gallery!",
	}
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, alert)
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
		case internal.ErrNotFound:
			vd.SetAlertErr("Invalid email address")
		case internal.ErrPasswordIncorrect:
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

type ResetPwForm struct {
	Email    string `schema:"email"`
	Token    string `schema:"token"`
	Password string `schema:"password"`
}

// POST /forgot
func (u *Users) InitiateReset(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPwForm
	vd.Yield = &form
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.ForgotPwView.Render(w, r, vd)
		return
	}

	token, err := u.us.InitiateReset(form.Email)
	if err != nil {
		vd.SetAlert(err)
		u.ForgotPwView.Render(w, r, vd)
		return
	}
	fmt.Println("Password reset token:", token)
	views.RedirectAlert(w, r, "/reset", http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Instructions for resetting the password have been sent to you.",
	})
}

// GET /reset
func (u *Users) ResetPw(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPwForm
	vd.Yield = &form

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.ResetPwView.Render(w, r, vd)
		return
	}
	u.ResetPwView.Render(w, r, vd)
}

// POST /reset
func (u *Users) CompleteReset(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPwForm
	vd.Yield = &form

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.ResetPwView.Render(w, r, vd)
		return
	}

	usr, err := u.us.CompleteReset(form.Token, form.Password)
	if err != nil {
		vd.SetAlert(err)
		u.ResetPwView.Render(w, r, vd)
		return
	}
	err = u.signIn(w, usr)
	if err != nil {
		vd.SetAlert(err)
		u.ResetPwView.Render(w, r, vd)
		return
	}
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Password successfully reset!",
	})
}

func (u *Users) signIn(w http.ResponseWriter, user *user.User) error {
	if user.Remember == "" {
		token, err := internal.GenerateRememberToken(internal.RememberTokenBytes)
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

func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	usr := context.User(r.Context())
	token, _ := internal.RememberToken()
	usr.Remember = token
	_ = u.us.Update(usr)
	http.Redirect(w, r, "/", http.StatusFound)
}
