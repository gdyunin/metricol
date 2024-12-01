/*
Package handlers provides HTTP handler functions for managing metrics.

This package includes handlers for sending error responses,
displaying metrics, and interacting with a storage repository.
*/
package handlers

import "net/http"

// BadRequest sends a response with status code 400 (Bad Request).
// It uses http.Error to write the response to the provided http.ResponseWriter.
func BadRequest(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

// NotFound sends a response with status code 404 (Not Found).
// It uses http.Error to write the response to the provided http.ResponseWriter.
func NotFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
