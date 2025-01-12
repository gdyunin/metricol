package general

import (
	"NewNewMetricol/internal/server/internal/entity"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type tr struct {
	Name  string
	Value string
}

type PullerAll interface {
	PullAll() (*entity.Metrics, error)
}

func MainPage(puller PullerAll) echo.HandlerFunc {
	return func(c echo.Context) error {
		allMetrics, err := puller.PullAll()
		if err != nil || allMetrics == nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		table := make([]*tr, 0, allMetrics.Length())
		for _, metric := range *allMetrics {
			name := metric.Name
			value := fmt.Sprint(metric.Value)
			table = append(table, &tr{
				Name:  name,
				Value: value,
			})
		}

		return c.Render(http.StatusOK, "main_page.html", table)
	}
}
