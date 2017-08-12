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
	"github.com/gorilla/mux"
)

type NotFoundError string

func (n NotFoundError) Error() string {
	return fmt.Sprintf("Entity : %s not found", n)
}

type Payment struct {
	Id              int64 `json:"-"`
	Amount          int64 `json:"amount"`
	PaidTo          string `json:"paidTo"`
	Day             time.Time `json:"day"`
	RemainingAmount int64 `json:"remainingBalance,omitempty"`
}

var connection = database.GetDataBaseConnection

func HandlePayment(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	personId := params["personId"]
	dairyId := params["dairyId"]

	var paymentDetails Payment
	err := decode(r, &paymentDetails)
	if err != nil {
		message := "Error while parsing payload to Payment reason :" + err.Error()
		log.Println(message)
		http.Error(w, message, http.StatusBadRequest)
	}

	paymentDetails.Day = time.Now()
	db := connection()

	err = updatePaymentDetails(dairyId, personId, &paymentDetails, db)
	if err != nil {
		switch err.(type) {
		case NotFoundError:
			http.Error(w, fmt.Sprintf("Person : %s does not exists", personId), http.StatusNotFound)
			break
		default:
			http.Error(w, fmt.Sprintf(err.Error()), http.StatusInternalServerError)
			break
		}
		return
	}

	err = encode(w, &paymentDetails)
	if err != nil {
		http.Error(w, "Error while encoding response", http.StatusInternalServerError)
	}
}

func GetPayments(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	personId := params["personId"]
	dairyId := params["dairyId"]

	db := connection()

	dairy := dairy.FindDairy(dairyId, db)
	if dairy == nil {
		log.Println("Dairy for the user does not exists ", dairyId)
		http.Error(w, fmt.Sprintf("Dairy: %s does not exists", dairyId), http.StatusNotFound)
		return
	}

	person := person.FindPerson(dairy.Id, personId, db)
	if person == nil {
		log.Println("Person does not exists ", personId)
		http.Error(w, fmt.Sprintf("Person: %s under Dairy: %s does not exists", personId, dairyId), http.StatusNotFound)
		return
	}

	payments, err := getPayments(dairy.Id, person.Id, db)
	if err != nil {
		message := "Error while reading all payments for person: "
		log.Println(message, personId, err)
		http.Error(w, fmt.Sprintf(message, personId), http.StatusInternalServerError)
	}

	err = encode(w, payments)
	if err != nil {
		message := "Error while writing payments to response"
		log.Println(message, err)
		http.Error(w, fmt.Sprintf(message), http.StatusInternalServerError)
	}
}

func getPayments(dairyRef, personRef int64, db *sql.DB) ([]Payment, error) {
	query := "SELECT * FROM payment_details where dairy_ref = $1 and person_ref = $2"
	rows, err := db.Query(query, dairyRef, personRef)
	if err != nil {
		return nil, err
	}

	var (
		payments          []Payment
		id, dairy, person int64
		amountPaid        int64
		paidTo            string
		day               time.Time
	)

	for rows.Next() {
		rows.Scan(&id, &dairy, &person, &amountPaid, &paidTo, &day)
		payments = append(payments, Payment{Amount: amountPaid, PaidTo: paidTo, Day: day})
	}

	return payments, nil
}

func updatePaymentDetails(dairyId, personId string, payment *Payment, db *sql.DB) error {
	dairy := dairy.FindDairy(dairyId, db)
	if dairy.Id == 0 {
		var a NotFoundError = "Dairy not found"
		return a
	}

	person := person.FindPerson(dairy.Id, personId, db)
	if person.Id == 0 {
		var a NotFoundError = "Person not found"
		return a
	}

	// insert payment
	err := insertPayment(dairy.Id, person.Id, *payment, db)
	if err != nil {
		log.Println("Erro while inserting payment details for person : ", personId)
		return err
	}

	// update balance
	remainingBalance, err := updateBalance(dairy.Id, person.Id, *payment, db)

	if err != nil {
		log.Println(fmt.Sprintf("Erro while updating remaining balance after payment : %d for person : %s", payment.Id, personId))
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
