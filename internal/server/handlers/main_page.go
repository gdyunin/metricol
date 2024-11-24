package handlers

import (
	"fmt"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"net/http"
)

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
	  <!-- ↓ INSERT TABLE ROWS WITH METRIC ↓ -->
      %s
    </table>
  </body>
</html>`

func MainPageHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var values string
		metricsAll := repository.Metrics()

		for _, t := range metricsAll {
			for k, v := range t {
				values += fmt.Sprintf("<tr><th>%s</th><th>%s</th></tr>", k, v)
			}
		}

		body := fmt.Sprintf(mainPageTemplate, values)

		w.Write([]byte(body))
		header := w.Header()
		header.Set("Content-Type", "text/html")
	}
}
