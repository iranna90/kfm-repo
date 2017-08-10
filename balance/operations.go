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
	DairyRef  int64
}

var connection = database.GetDataBaseConnection

func CreateRecord(personRef int64) error {
	db := connection()
	query := "INSERT INTO total_balance(person_ref, amount, last_updated) VALUES ($1,$2,$3)"
	_, err := db.Exec(query, personRef, 0, time.Now())
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
