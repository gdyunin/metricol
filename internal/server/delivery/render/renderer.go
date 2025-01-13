package render

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// Renderer is responsible for rendering HTML templates.
type Renderer struct {
	templates *template.Template // Parsed HTML templates.
}

// NewRenderer creates a new instance of Renderer.
//
// Parameters:
//   - templates: A pointer to a parsed set of templates.
//
// Returns:
//   - A new instance of Renderer.
func NewRenderer(templates *template.Template) *Renderer {
	return &Renderer{templates: templates}
}

// Render renders a template with the given name and data, writing the output to the provided writer.
//
// Parameters:
//   - w: An io.Writer where the rendered output will be written.
//   - name: The name of the template to render.
//   - data: The data to inject into the template.
//   - _: An echo.Context for compatibility, but it is not used in this implementation.
//
// Returns:
//   - An error if rendering fails; otherwise, nil.
func (t *Renderer) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	if err := t.templates.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("template rendering failed for template '%s' with data '%v': %w", name, data, err)
	}
	return nil
}
