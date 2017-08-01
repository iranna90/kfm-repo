package balance

import (
	"time"
	"kmf-repo/database"
	"database/sql"
)

type Balance struct {
	Id        int64
	Amount    int64
	Modified  time.Time
	PersonRef int64
}

var connection = database.GetDataBaseConnection

func CreateRecord(personRef int64) error {
	db := connection()
	query := "INSERT INTO total_balance(person_ref, amount, last_updated) VALUES ($1,$2,$3)"
	_, err := db.Exec(query, personRef, 0, time.Now())
	return err
}

func AddAmountToTotalBalance(personRef int64, todaysBalance int, db *sql.DB) (int64, error) {
	query := "UPDATE total_balance SET amount = amount + $1 where person_ref = $2 RETURNING amount"
	var newBalance int64
	err := db.QueryRow(query, todaysBalance, personRef).Scan(&newBalance)
	return newBalance, err
}

func RemovePayedAmountFromTotalBalance(personRef int64, todaysBalance int, db *sql.DB) (int64, error) {
	query := "UPDATE total_balance SET amount = amount - $1 where person_ref = $2 RETURNING amount"
	var remainingBalance int64
	err := db.QueryRow(query, todaysBalance, personRef).Scan(&remainingBalance)
	return remainingBalance, err
}
