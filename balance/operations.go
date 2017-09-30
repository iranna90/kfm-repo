package balance

import (
	"time"
	"kmf-repo/database"
	"database/sql"
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"fmt"
	"encoding/json"
)

type Balance struct {
	Id       int64 `json:"-"`
	Amount   int64 `json:"amount"`
	Modified time.Time `json:"lastUpdated"`
}

type PersonBalance struct {
	PersonId  string `json:"personId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Amount    int64 `json:"amount"`
}

var connection = database.GetDataBaseConnection

func CreateRecord(dairyRef int64, personRef int64) error {
	db := connection()
	query := "INSERT INTO total_balance(dairy_ref,person_ref, amount, last_updated) VALUES ($1,$2,$3,$4)"
	_, err := db.Exec(query, dairyRef, personRef, 0, time.Now())
	return err
}

func AddAmountToTotalBalance(dairyRef, personRef int64, todaysBalance int, db *sql.DB) (int64, error) {
	query := "UPDATE total_balance SET amount = amount + $1, last_updated = $2 where person_ref = $3 and dairy_ref = $4 RETURNING amount"
	var newBalance int64
	err := db.QueryRow(query, todaysBalance, time.Now(), personRef, dairyRef).Scan(&newBalance)
	return newBalance, err
}

func RemovePayedAmountFromTotalBalance(dairyRef, personRef int64, todaysBalance int, db *sql.DB) (int64, error) {
	query := "UPDATE total_balance SET amount = amount - $1 , last_updated = $2 where person_ref = $3 and dairy_Ref = $4 RETURNING amount"
	var remainingBalance int64
	err := db.QueryRow(query, todaysBalance, time.Now(), personRef, dairyRef).Scan(&remainingBalance)
	return remainingBalance, err
}

func GetBalance(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	personId := params["personId"]
	dairyId := params["dairyId"]
	db := connection()
	query := "SELECT amount, last_updated " +
		"from total_balance " +
		"where dairy_ref = (SELECT id from dairy d WHERE d.dairy_id = $1) " +
		"and person_ref = (SELECT id FROM persons p WHERE p.person_id = $2)"
	var (
		amount      int64
		lastUpdated time.Time
	)

	err := db.QueryRow(query, dairyId, personId).Scan(&amount, &lastUpdated)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error while reading balance record for person: %s", personId), http.StatusInternalServerError)
		return
	}

	balance := Balance{Amount: amount, Modified: lastUpdated}

	err = encode(w, &balance)
	if err != nil {
		log.Println("Erorr while writing response ", err)
		http.Error(w, fmt.Sprintf("Error while writing response for person: %s", personId), http.StatusInternalServerError)
	}
}

func GetPersonsBalance(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dairyId := params["dairyId"]
	queryString := r.URL.Query()["query"]
	var rows *sql.Rows
	var err error
	db := connection()
	if queryString != nil || len(queryString) > 0 {
		query := "SELECT p.person_id,p.first_name,p.last_name,b.amount FROM persons p INNER JOIN total_balance b " +
			"ON p.dairy_ref = (SELECT id FROM dairy d WHERE d.dairy_id = $1) " +
			"AND p.id = b.person_ref " +
			"AND (" +
			"p.person_id LIKE '%" + queryString[0] + "%' " +
			"OR p.first_name LIKE '%" + queryString[0] + "%' " +
			"OR p.last_name LIKE '%" + queryString[0] + "%'" +
			") " +
			"ORDER BY p.last_updated DESC;"
		rows, err = db.Query(query, dairyId)

	} else {
		limit := r.URL.Query()["limit"]
		offset := r.URL.Query()["offset"]
		if limit == nil || len(limit) == 0 || offset == nil || len(offset) == 0 {
			http.Error(w, "Unable to retrieve limit or offset from the request", http.StatusBadRequest)
			return
		}

		query := "select p.person_id, p.first_name, p.last_name, b.amount " +
			"from persons p inner join total_balance b " +
			"on p.dairy_ref = (SELECT id from dairy d WHERE d.dairy_id = $1) " +
			"and p.id = b.person_ref " +
			"ORDER BY  p.last_updated DESC LIMIT $2 OFFSET $3"

		rows, err = db.Query(query, dairyId, limit[0], offset[0])
	}

	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("Error while reading balance records for dairy: %s", dairyId), http.StatusInternalServerError)
		return
	}

	var (
		personDetails                 []PersonBalance
		amount                        int64
		personId, firstName, lastName string
	)
	for rows.Next() {
		rows.Scan(&personId, &firstName, &lastName, &amount)
		personDetails = append(personDetails, PersonBalance{PersonId: personId, FirstName: firstName, LastName: lastName, Amount: amount})
	}

	if len(personDetails) == 0 {
		personDetails = make([]PersonBalance, 0)
	}

	err = encode(w, personDetails)
	if err != nil {
		log.Println("Erorr while writing response ", err)
		http.Error(w, fmt.Sprintf("Error while writing response for dairy: %s", dairyId), http.StatusInternalServerError)
	}
}

func encode(w http.ResponseWriter, data interface{}) (err error) {
	encoder := json.NewEncoder(w)
	err = encoder.Encode(data)
	return
}
