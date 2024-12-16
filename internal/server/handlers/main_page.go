/*
Package handlers provides HTTP handler functions for managing metrics.

This package includes handlers for sending error responses,
displaying metrics, and interacting with a storage repository.
*/
package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gdyunin/metricol.git/internal/server/storage"
)

const mainPageTemplatePath = "web/template/main_page.html"

var cachedTemplate *template.Template

type TableRow struct {
	Name  string
	Value string
}

// MainPageHandler returns an HTTP handler function that generates
// an HTML page displaying metrics from the provided repository.
func MainPageHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsTable := fillMetricsTable(repository)

		t, err := parseTemplate(mainPageTemplatePath)
		if err != nil {
			log.Printf("failure getting template: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")

		if err := t.Execute(w, metricsTable); err != nil {
			log.Printf("failure executing template: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		// Response with response code 200 OK.
		w.WriteHeader(http.StatusOK)
	}
}

// fillMetricsTable constructs a slice of TableRow from the metrics
// stored in the provided repository.
func fillMetricsTable(repository storage.Repository) []TableRow {
	// Initialize a slice of TableRow with an initial length of 0 and a capacity
	// equal to the number of metrics in the repository. This avoids
	// multiple allocations as we append elements to the slice later.
	body := make([]TableRow, 0, repository.MetricsCount())

	metricsAll := repository.Metrics()
	for _, metricMap := range metricsAll {
		for metricName, metricValue := range metricMap {
			body = append(body, TableRow{
				Name:  metricName,
				Value: metricValue,
			})
		}
	}

	return body
}

// parseTemplate parses the main page HTML template and caches it for future use.
func parseTemplate(path string) (*template.Template, error) {
	// If the template was previously cached, return the cached version
	if cachedTemplate != nil {
		return cachedTemplate, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		// This is sad :(
		// TODO: In the future, if os.Getwd doesn't work, find another way to get the path to the template
		// to avoid breaking the function at this point.
		return nil, fmt.Errorf("error getting current working directory: %w", err)
	}
	cachedTemplate, err = template.ParseFiles(filepath.Join(currentDir, path))
	if err != nil {
		cachedTemplate = nil
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	return cachedTemplate, nil
}
