package general

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/labstack/echo/v4"
)

type tr struct {
	Name  string // The name or ID of the metric.
	Value string // The string representation of the metric's value.
}

func MainPage(adp *adapter.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		allMetrics, err := adp.PullAllMetrics()
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		table := make([]*tr, 0, len(allMetrics))
		for _, metric := range allMetrics {
			n := metric.ID
			v, err := metric.StringValue()
			if err != nil {
				v = "<invalid metric value>"
			}
			table = append(table, &tr{
				Name:  n,
				Value: v,
			})
		}

		return c.Render(http.StatusOK, "main_page.html", table)
	}
}
