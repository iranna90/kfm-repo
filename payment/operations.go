package payment

import (
	"time"
	"net/http"
	"encoding/json"
	"log"
	"kmf-repo/person"
	"kmf-repo/database"
	"database/sql"
)

type Payment struct {
	Id              int64 `json:"-"`
	PersonId        string `json:"personId"`
	amount          int64 `json:"amount"`
	PaidTo          string `json:"paidTo"`
	Day             time.Time `json:"day"`
	RemainingAmount int64 `json:"remainingBalance"`
}

var connection = database.GetDataBaseConnection

func HandlePayment(w http.ResponseWriter, r *http.Request) {
	var paymentDetails Payment
	err := decode(r, &paymentDetails)
	if err != nil {
		message := "Error while parsing payload to Payment reason :" + err.Error()
		log.Println(message)
		http.Error(w, message, http.StatusBadRequest)
	}
	db := connection()
	err = updatePaymentDetails(paymentDetails, db)
}

func updatePaymentDetails(payment Payment, db *sql.DB) error {
	person.FindPerson(payment.PersonId, db)
}

func decode(r *http.Request, dataType interface{}) (err error) {
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(dataType)
	return
}

func encode(w http.ResponseWriter, data interface{}) (err error) {
	encoder := json.NewEncoder(w)
	err = encoder.Encode(data)
	return
}