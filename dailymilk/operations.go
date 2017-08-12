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
	"kmf-repo/balance"
	"kmf-repo/dairy"
	"github.com/gorilla/mux"
)

type DailyMilkTransaction struct {
	Id              int64 `json:"-"`
	NumberOfLiters  int8 `json:"numberOfLiters"`
	TotalPriceOfDay int `json:"totalPriceOfTheDay"`
	Balance         int64 `json:"balance,omitempty"`
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

	params := mux.Vars(r)
	personId := params["personId"]
	dairyId := params["dairyId"]

	var transactionDetails DailyMilkTransaction
	err := decode(r, &transactionDetails)

	if err != nil {
		log.Println("Error while parsing request", err)
		http.Error(w, fmt.Sprintf("Invalid json: Unable to parse %s", err.Error()), http.StatusBadRequest)
		return
	}

	// insert the transaction details
	db := connection()
	err = updateTransaction(dairyId, personId, &transactionDetails, db)
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

func GetAllTransaction(w http.ResponseWriter, r *http.Request) {
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

	transactions, err := getAllTransaction(dairy.Id, person.Id, db)
	if err != nil {
		message := "Error while reading all transactions for person: "
		log.Println(message, personId, err)
		http.Error(w, fmt.Sprintf(message, personId), http.StatusInternalServerError)
	}

	err = encode(w, transactions)
	if err != nil {
		message := "Error while writing transaction to response"
		log.Println(message, err)
		http.Error(w, fmt.Sprintf(message), http.StatusInternalServerError)
	}
}

func getAllTransaction(dairyRef, personRef int64, connection *sql.DB) ([]DailyMilkTransaction, error) {
	query := "SELECT * FROM daily_transactions where dairy_ref = $1 and person_ref = $2"
	rows, err := connection.Query(query, dairyRef, personRef)
	if err != nil {
		return nil, err
	}

	var (
		transactions      []DailyMilkTransaction
		id, dairy, person int64
		numberOfListers   int8
		totalPrice        int
		day               time.Time
		personName        string
	)

	for rows.Next() {
		rows.Scan(&id, &dairy, &person, &numberOfListers, &totalPrice, &day, &personName)
		transactions = append(transactions, DailyMilkTransaction{NumberOfLiters: numberOfListers, TotalPriceOfDay: totalPrice, Day: day, PersonName: personName})
	}

	return transactions, nil
}

func updateTransaction(dairyId, personId string, transaction *DailyMilkTransaction, db *sql.DB, ) (err error) {
	// insert record
	err = calculatePriceOfMilk(transaction)
	if err != nil {
		return
	}

	dairy := dairy.FindDairy(dairyId, db)
	if dairy == nil {
		err = TransactionError{dairyId, "Dairy does not exitsts"}
		return
	}

	person := person.FindPerson(dairy.Id, personId, db)
	if person == nil {
		err = TransactionError{personId, "Person does not exitsts"}
		return
	}

	transaction.Day = time.Now()
	err = insertTransaction(dairy.Id, person.Id, transaction, db)
	if err != nil {
		return
	}

	// update total balance
	balance, err := balance.AddAmountToTotalBalance(dairy.Id, person.Id, transaction.TotalPriceOfDay, db)
	if err != nil {
		// TODO : Define user error
		return
	}
	transaction.Balance = int64(balance)

	return
}

func insertTransaction(dairyRef, personRef int64, transaction *DailyMilkTransaction, db *sql.DB) (err error) {
	query := "INSERT INTO daily_transactions(dairy_ref, person_ref, number_of_liters, total_price_of_day, day, person_name) VALUES ($1,$2,$3,$4,$5, $6)"
	_, err = db.Exec(query, dairyRef, personRef, transaction.NumberOfLiters, transaction.TotalPriceOfDay, transaction.Day, transaction.PersonName)
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
