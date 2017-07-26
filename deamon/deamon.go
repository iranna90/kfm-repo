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
	route.HandleFunc("/kmf/persons/{personId}", person.HandleGetPerson).Methods("GET")
	log.Println("Starting server")
	http.ListenAndServe(":1234", route)
}
