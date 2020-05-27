package main

import (
	"fmt"
	"net/http"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "Not found")
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintf(w, "Not found")
}

func PermissionDenied(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	fmt.Fprintf(w, "Not found")
}

func InvalidMethodHandler(w http.ResponseWriter, r *http.Request, allowed string) {
	w.Header().Set("Allow", allowed)
	w.WriteHeader(405)
	fmt.Fprintf(w, "Invalid method")
	return
}
