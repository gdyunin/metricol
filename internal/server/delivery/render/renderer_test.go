package render

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderer_Render(t *testing.T) {
	templates := template.Must(template.New("test").Parse(`{{.Title}}`))
	renderer := NewRenderer(templates)

	tests := []struct {
		name        string
		template    string
		data        interface{}
		expectError bool
		expected    string
	}{
		{
			name:        "Valid template rendering",
			template:    "test",
			data:        map[string]string{"Title": "Hello, world!"},
			expectError: false,
			expected:    "Hello, world!",
		},
		{
			name:        "Invalid template name",
			template:    "unknown",
			data:        map[string]string{"Title": "Hello"},
			expectError: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := renderer.Render(&buf, tt.template, tt.data, nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, buf.String())
			}
		})
	}
}
