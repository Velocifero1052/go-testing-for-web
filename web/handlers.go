package main

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"testingCourserWeb/pkg/data"
	"time"
)

var pathToTemplates = "./templates/"

type TemplateData struct {
	IP    string
	Data  map[string]any
	Error string
	Flash string
	User  data.User
}

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var td = make(map[string]any)
	if app.Session.Exists(r.Context(), "test") {
		msg := app.Session.GetString(r.Context(), "test")
		td["test"] = msg
	} else {
		app.Session.Put(r.Context(), "test", "Hit this page at "+time.Now().UTC().String())
	}
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{Data: td})
}

func (app *application) render(w http.ResponseWriter, r *http.Request, t string, td *TemplateData) error {
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
	}
	td.IP = app.ipFromContext(r.Context())
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Flash = app.Session.PopString(r.Context(), "flash")

	err = parsedTemplate.Execute(w, td)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "bad request", http.StatusBadRequest)
	}
	//validate data
	form := NewForm(r.PostForm)
	form.Required("email", "password")
	if !form.Valid() {
		//redirect to the login page with error message
		app.Session.Put(r.Context(), "error", "Invalid Login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		//redirect to the login page with error
		app.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// authenticate the user
	// if not authenticated then redirect with error
	if !app.authenticate(r, user, password) {
		app.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// prevent fixation attack
	_ = app.Session.RenewToken(r.Context())

	// store success message in session

	//redirect to some other page
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

func (app *application) authenticate(r *http.Request, user *data.User, password string) bool {
	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		return false
	}
	app.Session.Put(r.Context(), "user", user)
	return true
}

func (app *application) Profile(w http.ResponseWriter, r *http.Request) {
	_ = app.render(w, r, "profile.page.gohtml", &TemplateData{})
}
