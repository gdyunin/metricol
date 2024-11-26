package handlers

import "net/http"

// BadRequest send response with status code 400 (Bad Request).
func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

// NotFound send response with status code 404 (Not Found).
func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
