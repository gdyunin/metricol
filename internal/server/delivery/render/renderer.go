package render

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Renderer struct {
	templates *template.Template
}

func NewRenderer(templates *template.Template) *Renderer {
	return &Renderer{templates: templates}
}

func (t *Renderer) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	if err := t.templates.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("failed to execute render '%s': %w", name, err)
	}
	return nil
}
