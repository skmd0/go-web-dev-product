package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func NewView(layout string, layouts ...string) (*View, error) {
	prependViewDir(layouts)
	appendGoHTMLExt(layouts)
	layouts = append(layouts, getLayouts()...)
	t, err := template.ParseFiles(layouts...)
	if err != nil {
		return nil, err
	}
	return &View{
		Template: t,
		Layout:   layout,
	}, nil
}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := v.Render(w, r); err != nil {
		panic(err)
	}
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

func getLayouts() []string {
	layouts, err := filepath.Glob("views/layouts/*.gohtml")
	if err != nil {
		panic(err)
	}
	return layouts
}

func prependViewDir(layouts []string) {
	for i, l := range layouts {
		layouts[i] = "views/" + l
	}
}

func appendGoHTMLExt(layouts []string) {
	for i, l := range layouts {
		layouts[i] = l + ".gohtml"
	}
}
