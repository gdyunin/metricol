package handlers

import (
	"bytes"
	"fmt"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"net/http"
)

// Full page template
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

// One table row template
const rowTemplate = "<tr><th>%s</th><th>%s</th></tr>"

// MainPageHandler return a handler that generates a page with known metrics
func MainPageHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buffer bytes.Buffer

		// Pull known metrics and fill body
		metricsAll := repository.Metrics()
		for _, metricMap := range metricsAll {
			for metricName, metricValue := range metricMap {
				buffer.WriteString(fmt.Sprintf(rowTemplate, metricName, metricValue))
			}
		}
		body := fmt.Sprintf(mainPageTemplate, buffer.String())

		// The error is ignored as it has no effect
		// A logger could be added in the future
		_, _ = w.Write([]byte(body))

		w.Header().Set("Content-Type", "text/html")
	}
}
