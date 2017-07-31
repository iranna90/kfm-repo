package dailymilk

import (
	"time"
	"net/http"
	"kmf-repo/database"
	"encoding/json"
	"log"
	"fmt"
	"database/sql"
	"kmf-repo/person"
	balance2 "kmf-repo/balance"
)

type DailyMilkTransaction struct {
	Id              int64 `json:"-"`
	PersonId        string `json:"personId"`
	NumberOfLiters  int8 `json:"numberOfLiters"`
	TotalPriceOfDay int `json:"totalPriceOfTheDay"`
	Balance         int64 `json:"balance"`
	Day             time.Time `json:"time"`
	PersonName      string `json:"personName"`
}

type TransactionError struct {
	personId string
	message  string
}

func (t TransactionError) Error() string {
	return fmt.Sprintf("Error while doing operation for person : %s and error details are %s", t.personId, t.message)
}

var pricePerLiter = 41

var connection = database.GetDataBaseConnection

func HandleMilkSubmission(w http.ResponseWriter, r *http.Request) {

	var transactionDetails DailyMilkTransaction
	err := decode(r, &transactionDetails)

	if err != nil {
		log.Println("Error while parsing request", err)
		http.Error(w, fmt.Sprintf("Invalid json: Unable to parse %s", err.Error()), http.StatusBadRequest)
		return
	}

	// insert the transaction details
	db := connection()
	err = updateTransaction(db, &transactionDetails)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro while updating balance: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	err = encode(w, transactionDetails)
	if err != nil {
		http.Error(w, "Unable to encode response to json", http.StatusInternalServerError)
	}
	log.Println("Successfully updated the balance")
}

func updateTransaction(db *sql.DB, transaction *DailyMilkTransaction) (err error) {
	// insert record
	err = calculatePriceOfMilk(transaction)
	if err != nil {
		return
	}

	person := person.FindPerson(transaction.PersonId, db)
	if person == nil {
		err = TransactionError{transaction.PersonId, "Person does not exitsts"}
		return
	}
	transaction.Day = time.Now()
	insertTransaction(person.Id, transaction, db)

	// update total balance
	balance, err := balance2.UpdateTotalBalance(person.Id, transaction.TotalPriceOfDay, db)
	if err != nil {
		// TODO : Define user error
		return
	}
	transaction.Balance = int64(balance)

	return
}

func insertTransaction(personRef int64, transaction *DailyMilkTransaction, db *sql.DB) (err error) {
	query := "INSERT INTO daily_transactions(person_ref, number_of_liters, total_price_of_day, day, person_name) VALUES ($1,$2,$3,$4,$5)"
	_, err = db.Exec(query, personRef, transaction.NumberOfLiters, transaction.TotalPriceOfDay, transaction.Day, transaction.PersonName)
	return
}

type TotalCalculationError string

func (t TotalCalculationError) Error() string {
	return t.Error()
}

func calculatePriceOfMilk(transaction *DailyMilkTransaction) (err error) {
	if transaction.NumberOfLiters <= 0 {
		var t TotalCalculationError = "Number of liters should always be greater then 0"
		err = t
		return err
	}
	transaction.TotalPriceOfDay = int(transaction.NumberOfLiters) * pricePerLiter
	return
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
