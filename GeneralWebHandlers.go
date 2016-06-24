package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// WebIndex handles requests to /
func WebIndex(w http.ResponseWriter, r *http.Request) {
	if HSTSENABLED == true {
		w.Header().Set("Strict-Transport-Security", "max-age="+strconv.FormatInt(HSTSEXPIRY, 10)+"; includeSubdomains")
	}

	tmpl, err := template.New("index").ParseFiles("assets/templates/index.html")
	err = tmpl.Execute(w, nil)

	if err != nil {
		log.Fatal(err)
	}

}

// WebAbout handles requests to /about/
func WebAbout(w http.ResponseWriter, r *http.Request) {
	if HSTSENABLED == true {
		w.Header().Set("Strict-Transport-Security", "max-age="+strconv.FormatInt(HSTSEXPIRY, 10)+"; includeSubdomains")
	}

	tmpl, err := template.New("about").ParseFiles("assets/templates/about.html")

	err = tmpl.Execute(w, nil)

	if err != nil {
		log.Fatal(err)
	}

}
