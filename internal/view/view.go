package view

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
)

//go:embed all:*.html email/*.html
var templatesFS embed.FS

type EmailData struct {
	FullName string
	URL      string
}

func Render(w http.ResponseWriter, contentFile string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFS(templatesFS, "layout.html", contentFile)
	if err != nil {
		http.Error(w, "internal server error: failed to parse web templates", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "internal server error: failed to render web view", http.StatusInternalServerError)
		return
	}
}

func RenderEmail(templateName string, data EmailData) (string, error) {
	tmpl, err := template.ParseFS(templatesFS, "email/"+templateName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
