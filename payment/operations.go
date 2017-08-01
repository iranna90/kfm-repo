package payment

import (
	"time"
	"net/http"
	"encoding/json"
	"log"
	"kmf-repo/person"
	"kmf-repo/database"
	"database/sql"
	"fmt"
	"kmf-repo/balance"
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

	paymentDetails.Day = time.Now()
	db := connection()

	err = updatePaymentDetails(&paymentDetails, db)
	if err != nil {
		switch err.(type) {
		case person.PersonError:
			http.Error(w, fmt.Sprintf("Person : %s does not exists", paymentDetails.PersonId), http.StatusNotFound)
			break
		default:
			http.Error(w, fmt.Sprintf(err.Error()), http.StatusInternalServerError)
		}
	}

	err = encode(w, &paymentDetails)
	if err != nil {
		http.Error(w, "Error while encoding response", http.StatusInternalServerError)
	}
}

func updatePaymentDetails(payment *Payment, db *sql.DB) error {

	person := person.FindPerson(payment.PersonId, db)
	if person.Id == 0 {
		return person.PersonError{payment.PersonId, "Not found"}
	}

	// insert payment
	err := insertPayment(person.Id, *payment, db)
	if err != nil {
		log.Println("Erro while inserting payment details for person : ", person.PersonId)
		return err
	}

	// update balance
	remainingBalance, err := updateBalance(person.Id, *payment, db)

	if err != nil {
		log.Println(fmt.Sprintf("Erro while updating remaining balance after payment : %d for person : %s", payment.Id, person.PersonId))
		return err
	}

	payment.RemainingAmount = remainingBalance
	return nil
}

func updateBalance(personRef int64, payment Payment, db *sql.DB) (remainingBalance int64, err error) {
	remainingBalance, err = balance.RemovePayedAmountFromTotalBalance(personRef, payment.amount, db)
	return
}

func insertPayment(personRef int64, payment Payment, db *sql.DB) error {
	query := "INSERT INTO payment_details(person_ref, amount_payed, paid_to, day) VALUES ($1,$2,$3,$4)"
	_, err := db.Exec(query, personRef, payment.amount, payment.PaidTo, payment)
	return err
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
