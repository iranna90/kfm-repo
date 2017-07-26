package main

import (
	"github.com/gorilla/mux"
	"kmf-repo/person"
	"net/http"
	"log"
)

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/kmf/persons/{personId}", person.HandlerCreateOrUpdatePerson).Methods("PUT")
	log.Println("Starting server")
	http.ListenAndServe(":1234", route)
}
