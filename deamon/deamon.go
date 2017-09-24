package main

import (
	"github.com/gorilla/mux"
	"kmf-repo/person"
	"net/http"
	"log"
	"kmf-repo/dailymilk"
	"kmf-repo/dairy"
	"kmf-repo/balance"
)

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/kmf/dairies/{dairyId}", dairy.HandlerCreateOrUpdateDairy).Methods("POST")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}", person.HandlerCreateOrUpdatePerson).Methods("POST")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}", person.HandleGetPerson).Methods("GET")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}/balance", balance.GetBalance).Methods("GET")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons", balance.GetPersonsBalance).Methods("GET")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}/transactions", dailymilk.HandleMilkSubmission).Methods("POST")
	route.HandleFunc("/kmf/dairies/{dairyId}/persons/{personId}/transactions", dailymilk.GetAllTransaction).Methods("GET")
	route.HandleFunc("/kmf/dairies/{dairyId}/transactions", dailymilk.GetAllTransactionOfDairy).Methods("GET")
	log.Println("Starting server")
	http.ListenAndServe(":1234", route)
}
