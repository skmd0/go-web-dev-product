package views

import (
	"bytes"
	"go-web-dev/context"
	"html/template"
	"io"
	"log"
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
	v.Render(w, r, nil)
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{Yield: data}
	}
	vd.User = context.User(r.Context())

	var buf bytes.Buffer
	err := v.Template.ExecuteTemplate(&buf, v.Layout, vd)
	if err != nil {
		log.Println(err)
		http.Error(w, AlertMsgGeneric, http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, &buf)
	if err != nil {
		http.Error(w, AlertMsgGeneric, http.StatusInternalServerError)
		return
	}
}

func getLayouts() []string {
	layouts, err := filepath.Glob("../views/layouts/*.gohtml")
	if err != nil {
		panic(err)
	}
	return layouts
}

func prependViewDir(layouts []string) {
	for i, l := range layouts {
		layouts[i] = "../views/" + l
	}
}

func appendGoHTMLExt(layouts []string) {
	for i, l := range layouts {
		layouts[i] = l + ".gohtml"
	}
}
