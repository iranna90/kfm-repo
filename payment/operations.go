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
	"kmf-repo/dairy"
)

type NotFoundError string

func (n NotFoundError) Error() string {
	return fmt.Sprintf("Entity : %s not found", n)
}

type Payment struct {
	Id              int64 `json:"-"`
	DairyId         string `json:"dairyId"`
	PersonId        string `json:"personId"`
	Amount          int64 `json:"amount"`
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
		case NotFoundError:
			http.Error(w, fmt.Sprintf("Person : %s does not exists", paymentDetails.PersonId), http.StatusNotFound)
			break
		default:
			http.Error(w, fmt.Sprintf(err.Error()), http.StatusInternalServerError)
			break
		}
	}

	err = encode(w, &paymentDetails)
	if err != nil {
		http.Error(w, "Error while encoding response", http.StatusInternalServerError)
	}
}

func updatePaymentDetails(payment *Payment, db *sql.DB) error {
	dairy := dairy.FindDairy(payment.DairyId, db)
	if dairy.Id == 0 {
		var a NotFoundError = "Dairy not found"
		return a
	}

	person := person.FindPerson(dairy.Id, payment.PersonId, db)
	if person.Id == 0 {
		var a NotFoundError = "Person not found"
		return a
	}

	// insert payment
	err := insertPayment(dairy.Id, person.Id, *payment, db)
	if err != nil {
		log.Println("Erro while inserting payment details for person : ", person.PersonId)
		return err
	}

	// update balance
	remainingBalance, err := updateBalance(dairy.Id, person.Id, *payment, db)

	if err != nil {
		log.Println(fmt.Sprintf("Erro while updating remaining balance after payment : %d for person : %s", payment.Id, person.PersonId))
		return err
	}

	payment.RemainingAmount = remainingBalance
	return nil
}

func updateBalance(dairyRef int64, personRef int64, payment Payment, db *sql.DB) (remainingBalance int64, err error) {
	remainingBalance, err = balance.RemovePayedAmountFromTotalBalance(dairyRef, personRef, payment.Amount, db)
	return
}

func insertPayment(dairyRef int64, personRef int64, payment Payment, db *sql.DB) error {
	query := "INSERT INTO payment_details(dairy_ref, person_ref, amount_payed, paid_to, day) VALUES ($1,$2,$3,$4, $5)"
	_, err := db.Exec(query, dairyRef, personRef, payment.Amount, payment.PaidTo, payment.Day)
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
