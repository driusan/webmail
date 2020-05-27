package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

var templates *template.Template

func emptyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func cacheUpdater() {
	for {
		time.Sleep(2 * time.Minute)
		loadcache()
	}
}
func main() {
	templ, err := template.ParseFiles(
		"templates/index.html",
		"templates/message.html",
		"templates/login.html",
		"templates/new.html",
	)
	if err != nil {
		log.Fatalln(err)
	}
	templates = templ

	println(templates.DefinedTemplates())
	loadcache()

	http.HandleFunc("/messages/", MessageHandler)
	http.HandleFunc("/new", NewHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/favicon.ico", emptyHandler)
	println("Starting server")
	go cacheUpdater()
	if err := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil); err != nil {
		log.Fatalln(err)
	}
}
