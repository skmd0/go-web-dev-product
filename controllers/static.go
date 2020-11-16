package controllers

import "go-web-dev/views"

func NewStatic() (*StaticViews, error) {
	home, err := views.NewView("bulma", "static/home")
	if err != nil {
		return nil, err
	}
	contact, err := views.NewView("bulma", "static/contact")
	if err != nil {
		return nil, err
	}
	return &StaticViews{
		Home:    home,
		Contact: contact,
	}, nil
}

type StaticViews struct {
	Home    *views.View
	Contact *views.View
}
