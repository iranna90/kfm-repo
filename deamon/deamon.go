package main

import (
	"github.com/gorilla/mux"
	"kmf-repo/person"
	"net/http"
	"log"
	"kmf-repo/dailymilk"
	"kmf-repo/payment"
)

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/kmf/persons/{personId}", person.HandlerCreateOrUpdatePerson).Methods("PUT")
	route.HandleFunc("/kmf/persons/{personId}", person.HandleGetPerson).Methods("GET")
	route.HandleFunc("/kmf/persons/{personId}/transactions", dailymilk.HandleMilkSubmission).Methods("POST")
	route.HandleFunc("/kmf/persons/{personId}/payments", payment.HandlePayment).Methods("POST")
	log.Println("Starting server")
	http.ListenAndServe(":1234", route)
}
