package echohttp

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer is responsible for rendering HTML templates in an Echo framework context.
type TemplateRenderer struct {
	templates *template.Template // A parsed collection of templates.
}

// Render renders an HTML template and writes the output to the provided writer.
//
// Parameters:
//   - w: The writer to which the rendered template is written.
//   - name: The name of the template to render.
//   - data: The data to populate the template with.
//   - _: The Echo context (not used in this implementation).
//
// Returns:
//   - An error if the template execution fails, wrapped with additional context.
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	if err := t.templates.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("failed to execute template '%s': %w", name, err)
	}
	return nil
}
