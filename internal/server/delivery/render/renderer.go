// Package render provides an HTML template renderer for the Echo framework.
// It enables the rendering of dynamic HTML content by injecting data into pre-parsed templates.
package render

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// Renderer is responsible for rendering HTML templates.
// It holds a pointer to a set of parsed HTML templates that are used to generate the final output.
type Renderer struct {
	templates *template.Template // templates holds the parsed HTML templates.
}

// NewRenderer creates and returns a new Renderer instance.
//
// Parameters:
//   - templates: A pointer to a parsed set of HTML templates.
//
// Returns:
//   - *Renderer: A new instance of Renderer configured with the provided templates.
func NewRenderer(templates *template.Template) *Renderer {
	return &Renderer{templates: templates}
}

// Render renders a template with the given name and data, writing the output to the provided writer.
// This method implements the echo.Renderer interface, allowing it to be used as a custom renderer in Echo.
//
// Parameters:
//   - w: An io.Writer where the rendered output will be written.
//   - name: The name of the template to render.
//   - data: The data to inject into the template.
//   - _ : The echo.Context parameter is provided for compatibility with the echo.Renderer interface but is not used.
//
// Returns:
//   - error: An error if rendering fails; otherwise, nil.
func (t *Renderer) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	if err := t.templates.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("template rendering failed for template '%s' with data '%v': %w", name, data, err)
	}
	return nil
}
