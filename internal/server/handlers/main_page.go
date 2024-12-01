package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gdyunin/metricol.git/internal/server/storage"
)

// Full page template.
const mainPageTemplate = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Известные метрики</title>
    <style>
      body {
      background: #003366
      }
      table {
      font-family: "Lucida Sans Unicode", "Lucida Grande", Sans-Serif;
      font-size: 18px;
      border-collapse: collapse;
      text-align: center;
      }
      th, td:first-child {
      background: #AFCDE7;
      padding: 10px 20px;
      }
      th, td {
      border-style: solid;
      border-width: 0 1px 1px 0;
      border-color: white;
      }
      td {
      background: #D8E6F3;
      }
      th:first-child, td:first-child {
      text-align: left;
      }
    </style>
  </head>
  <body>
    <table>
      <tr>
        <th>Метрика</th>
        <th>Значение</th>
      </tr>
      %s <!-- rows with metrics will be here -->
    </table>
  </body>
</html>`

// Table row template.
const rowTemplate = "<tr><th>%s</th><th>%s</th></tr>"

// MainPageHandler return a handler that generates a page with known metrics.
func MainPageHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//var buffer bytes.Buffer
		type tableRow struct {
			Name  string
			Value string
		}
		body := []tableRow{}

		// Pull known metrics and fill body.
		metricsAll := repository.Metrics()
		for _, metricMap := range metricsAll {
			for metricName, metricValue := range metricMap {
				body = append(body, tableRow{
					Name:  metricName,
					Value: metricValue,
				})
				//buffer.WriteString(fmt.Sprintf(rowTemplate, metricName, metricValue))
			}
		}
		//_ = fmt.Sprintf(mainPageTemplate, buffer.String())

		currentDir, _ := os.Getwd()
		t, err := template.ParseFiles(filepath.Join(currentDir, "web", "template", "main_page.html"))
		t.Execute(w, body)

		fmt.Println(t)
		fmt.Println(err)

		// The error is ignored as it has no effect.
		// A logger could be added in the future.
		//_, _ = w.Write([]byte(body))

		w.Header().Set("Content-Type", "text/html")
	}
}
