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

func RemovePayedAmountFromTotalBalance(dairyRef, personRef int64, todaysBalance int64, db *sql.DB) (int64, error) {
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
	query := "SELECT amount, last_updated from total_balance where dairy_ref = (SELECT id from dairy d WHERE d.dairy_id = $1) and person_ref = (SELECT id FROM persons p WHERE p.person_id = $2)"
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

func encode(w http.ResponseWriter, data interface{}) (err error) {
	encoder := json.NewEncoder(w)
	err = encoder.Encode(data)
	return
}
