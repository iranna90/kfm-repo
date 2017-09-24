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
	"strings"
)

type DailyMilkTransaction struct {
	Id             int64 `json:"id"`
	NumberOfLiters int `json:"numberOfLiters"`
	Amount         int `json:"amount"`
	Balance        int64 `json:"closingBalance,omitempty"`
	Day            time.Time `json:"day"`
	Type           string `json:"transactionType"`
	PersonName     string `json:"personName"`
}

type DairyTransactions struct {
	DailyMilkTransaction
	PersonId  string `json:"personId,omitempty"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type TransactionError struct {
	personId string
	message  string
}

const (
	PAID    = "PAID"
	DEPOSIT = "DEPOSIT"
)

func (t TransactionError) Error() string {
	return fmt.Sprintf("Error while doing operation for person : %s and error details are %s", t.personId, t.message)
}

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
	err = newTransaction(dairyId, personId, &transactionDetails, db)
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
	limit := r.URL.Query()["limit"]
	offset := r.URL.Query()["offset"]

	if limit == nil || len(limit) == 0 || offset == nil || len(offset) == 0 {
		http.Error(w, "Unable to retrieve limit or offset from the request", http.StatusBadRequest)
		return
	}

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

	transactions, err := getAllTransaction(limit[0], offset[0], dairy.Id, person.Id, db)
	if err != nil {
		message := "Error while reading all transactions for person: "
		log.Println(message, personId, err)
		http.Error(w, fmt.Sprintf(message, personId), http.StatusInternalServerError)
	}

	if transactions == nil {
		transactions = make([]DailyMilkTransaction, 0)
	}

	err = encode(w, transactions)
	if err != nil {
		message := "Error while writing transaction to response"
		log.Println(message, err)
		http.Error(w, fmt.Sprintf(message), http.StatusInternalServerError)
	}
}

func GetAllTransactionOfDairy(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dairyId := params["dairyId"]

	from, err := time.Parse(time.RFC3339, strings.Replace(r.URL.Query().Get("from"), " ", "+", 1))
	if err != nil {
		message := "invalid From date, So please provide valid date as query parameters"
		log.Println(message, err)
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	to, err := time.Parse(time.RFC3339, strings.Replace(r.URL.Query().Get("to"), " ", "+", 1))
	if err != nil {
		message := "invalid To date , So please provide valid to date as query parameters"
		log.Println(message, err)
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	db := connection()

	dairy := dairy.FindDairy(dairyId, db)
	if dairy == nil {
		log.Println("Dairy for the user does not exists ", dairyId)
		http.Error(w, fmt.Sprintf("Dairy: %s does not exists", dairyId), http.StatusNotFound)
		return
	}

	query := "select p.person_id, p.first_name, p.last_name, tr.person_name, tr.number_of_liters, tr.total_price_of_day, tr.day, tr.transaction_type" +
		"from persons p inner join transactions tr " +
		"on p.dairy_ref = $1 " +
		"and p.id = tr.person_ref " +
		"and tr.day between $2 and  $3 order by tr.day DESC"

	var (
		personDetails                                          []DairyTransactions
		amount                                                 int
		liters                                                 int
		personId, firstName, lastName, paidTo, transactionType string
		date                                                   time.Time
	)

	rows, err := db.Query(query, dairy.Id, from, to)

	if err != nil {
		message := "Error while reading transactions records for dairy: %s, strat date: %s and end date: %s"
		log.Println(fmt.Sprintf(message, dairyId, from, to), err)
		http.Error(w, fmt.Sprintf(message, dairyId, from, to), http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		rows.Scan(&personId, &firstName, &lastName, &paidTo, &liters, &amount, &date, &transactionType)
		personDetails = append(personDetails, DairyTransactions{DailyMilkTransaction: DailyMilkTransaction{PersonName: paidTo, Amount: amount, NumberOfLiters: liters, Day: date, Type: transactionType}, PersonId: personId, FirstName: firstName, LastName: lastName})
	}

	if len(personDetails) == 0 {
		message := "No transactions found from %s to %s for dairy %s"
		log.Println(fmt.Sprintf(message, from, to, dairyId))
		http.Error(w, fmt.Sprintf(message, from, to, dairyId), http.StatusNotFound)
		return
	}

	err = encode(w, personDetails)
	if err != nil {
		message := "Erorr while writing response "
		log.Println(message, err)
		http.Error(w, fmt.Sprintf("Error while writing response for dairy: %s", dairyId), http.StatusInternalServerError)
	}

}

func getAllTransaction(limit, offset string, dairyRef, personRef int64, connection *sql.DB) ([]DailyMilkTransaction, error) {
	query := "SELECT * FROM transactions tr where dairy_ref = $1 and person_ref = $2 ORDER BY  tr.day DESC LIMIT $3 OFFSET $4"
	rows, err := connection.Query(query, dairyRef, personRef, limit, offset)
	if err != nil {
		return nil, err
	}

	var (
		transactions                       []DailyMilkTransaction
		id, dairy, person, remainingAmount int64
		numberOfListers                    int
		totalPrice                         int
		day                                time.Time
		personName, transactionType        string
	)

	for rows.Next() {
		rows.Scan(&id, &dairy, &person, &numberOfListers, &totalPrice, &remainingAmount, &day, &personName, &transactionType)
		transactions = append(transactions,
			DailyMilkTransaction{
				Id:             id,
				NumberOfLiters: numberOfListers,
				Amount:         totalPrice,
				Day:            day,
				PersonName:     personName,
				Type:           transactionType,
				Balance:        remainingAmount})
	}

	return transactions, nil
}

func newTransaction(dairyId, personId string, transaction *DailyMilkTransaction, db *sql.DB, ) (err error) {
	// insert record
	log.Println("transaction type is ", transaction.Type)
	if transaction.Type != PAID && transaction.Type != DEPOSIT {
		err = TransactionError{dairyId, "Only operation supported are PAID or DEPOSIT"}
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

	var totalBalance int64
	if transaction.Type == PAID {
		log.Println("Transaction type is payment for user ", personId, " Amount is:", transaction.Amount)
		totalBalance, err = balance.RemovePayedAmountFromTotalBalance(dairy.Id, person.Id, transaction.Amount, db)
	} else {
		log.Println("Transaction type is deposit for user ", personId, " Amount is:", transaction.Amount)
		totalBalance, err = balance.AddAmountToTotalBalance(dairy.Id, person.Id, transaction.Amount, db)
	}

	if err != nil {
		return
	}
	transaction.Balance = int64(totalBalance)
	transaction.Day = time.Now()
	id, err := insertTransaction(dairy.Id, person.Id, transaction, db)
	if err != nil {
		return
	}
	transaction.Id = id
	return
}

func insertTransaction(dairyRef, personRef int64, transaction *DailyMilkTransaction, db *sql.DB) (id int64, err error) {
	var transactionId int64
	query := "INSERT INTO transactions(dairy_ref, person_ref, number_of_liters, amount, remaining_total, day, person_name, transaction_type) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id"
	err = db.QueryRow(query, dairyRef, personRef, transaction.NumberOfLiters, transaction.Amount, transaction.Balance, transaction.Day, transaction.PersonName, transaction.Type).Scan(&transactionId)
	return transactionId, err
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
