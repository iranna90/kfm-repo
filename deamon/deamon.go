package main

import (
	"github.com/gorilla/mux"
	"kmf-repo/person"
	"net/http"
	"log"
	"kmf-repo/dailymilk"
	"kmf-repo/payment"
	"kmf-repo/dairy"
)

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/kmf/dairies/{dairyId}", dairy.HandlerCreateOrUpdateDairy)
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}", person.HandlerCreateOrUpdatePerson).Methods("PUT")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}", person.HandleGetPerson).Methods("GET")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}/transactions", dailymilk.HandleMilkSubmission).Methods("POST")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}/payments", payment.HandlePayment).Methods("POST")
	log.Println("Starting server")
	http.ListenAndServe(":1234", route)
}
