package web

import (
	"embed"
	"html/template"
	"io"
	"io/fs"

	"github.com/labstack/echo/v4"
)

//go:embed templates/* manifest.json
var templateFS embed.FS

// FS exposes the embedded filesystem for serving static files.
var FS = templateFS

// allLayouts are parsed for every render so sub-templates are always available.
var baseFiles = []string{
	"templates/layout.html",
	"templates/today.html",
	"templates/summary.html",
}

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	// htmx fragment responses — no layout wrapper
	fragments := map[string]string{
		"today.html":   "today.html",
		"summary.html": "summary.html",
	}

	fsys, _ := fs.Sub(templateFS, ".")

	if tmplName, ok := fragments[name]; ok {
		t, err := template.New("").ParseFS(fsys, "templates/today.html", "templates/summary.html")
		if err != nil {
			return err
		}
		return t.ExecuteTemplate(w, tmplName, data)
	}

	files := append(baseFiles, "templates/"+name)
	t, err := template.New("").ParseFS(fsys, files...)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, "layout.html", data)
}
