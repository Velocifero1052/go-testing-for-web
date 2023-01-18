package main

import (
	"html/template"
	"net/http"
)

type TemplateData struct {
	IP   string
	Data map[string]any
}

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{})
}

func (app *application) render(w http.ResponseWriter, r *http.Request, t string, data *TemplateData) error {
	parsedTemplate, err := template.ParseFiles("./templates/" + t)

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	err = parsedTemplate.Execute(w, data)
	if err != nil{
		return err
	}

	return nil
}
