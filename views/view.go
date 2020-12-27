package views

import (
	"bytes"
	"github.com/gorilla/csrf"
	"github.com/pkg/errors"
	"go-web-dev/internal"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

const (
	templatesDirPathGlob = "../views/layouts/*.gohtml"
)

func NewView(baseTemplate string, templates ...string) (*View, error) {
	prependViewDir(templates)
	appendGoHTMLExt(templates)
	partialTemplates, err := getPartialTemplates(templatesDirPathGlob)
	if err != nil {
		return nil, err
	}
	templates = append(templates, partialTemplates...)
	// csrf protection
	t, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("CSRF is not implemented")
		},
	}).ParseFiles(templates...)
	if err != nil {
		return nil, err
	}
	return &View{
		Template: t,
		Layout:   baseTemplate,
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
	if alert := getAlert(r); alert != nil && vd.Alert == nil {
		vd.Alert = alert
		clearAlert(w)
	}
	vd.User = internal.GetUser(r.Context())

	var buf bytes.Buffer
	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})
	err := tpl.ExecuteTemplate(&buf, v.Layout, vd)
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

func getPartialTemplates(path string) ([]string, error) {
	templates, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func prependViewDir(templates []string) {
	for i, l := range templates {
		templates[i] = "../views/" + l
	}
}

func appendGoHTMLExt(templates []string) {
	for i, l := range templates {
		templates[i] = l + ".gohtml"
	}
}
