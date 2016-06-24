package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"html/template"
	"log"
	"net/http"
)

func WebManage(w http.ResponseWriter, r *http.Request, db *bolt.DB, FQDN string) {

	if r.Method == "POST" {
		tmpl, err := template.New("create").ParseFiles("assets/templates/create.html")
		err = tmpl.Execute(w, nil)

		if err != nil {
			log.Fatal(err)
		}

	} else if r.Method == "GET" {

		tmpl, err := template.New("create").ParseFiles("assets/templates/create.html")
		err = tmpl.Execute(w, nil)

		if err != nil {
			log.Fatal(err)
		}

	} else {
		//We don't accept PUT / DELETE etc on the /create/ URL
		w.Header().Set("ow-success", "false")
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Method %s is not allowed!", r.Method)
		fmt.Fprintf(w, "Method %s is not allowed!", r.Method)
	}
}
