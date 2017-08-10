package dairy

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"kmf-repo/database"
	_ "github.com/lib/pq"
	. "database/sql"
)

type Dairy struct {
	Id      int64 `json:"-"`
	DairyId string `json:"dairyId"`
}

var dataBaseConnection = database.GetDataBaseConnection

func HandlerCreateOrUpdateDairy(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var dairy Dairy
	err := decode(r, &dairy)
	if err != nil {
		log.Println("Error while parsing payload to dairy", err)
		http.Error(w, fmt.Sprint("Error while parsing payload to dairy", err), http.StatusBadRequest)
		return
	}

	db := dataBaseConnection()

	retrievedDairy := FindDairy(dairy.DairyId, db)

	// check person exists
	if retrievedDairy.Id != 0 {
		// update
		log.Println("Dairy ", dairy.DairyId, "already exists,Update is not supported")
		http.Error(w, fmt.Sprintf("Dairy with id : %s already exists", dairy.DairyId), http.StatusConflict)
		return
	}

	err = insertDairyRecord(&dairy, db)
	if err != nil {
		log.Println("Error while inserting dairy details ", err)
		http.Error(w, fmt.Sprint("Error while writing to data base", err), http.StatusInternalServerError)
		return
	}

	err = encode(w, dairy)
	if err != nil {
		log.Println("Error while writing the payload", err)
		http.Error(w, fmt.Sprint("Error while writing the payload", err), http.StatusInternalServerError)
		return
	}

	log.Println("Dairy careted successfully")
}

func insertDairyRecord(dairy *Dairy, db *DB) (error) {
	var id int64
	err := db.QueryRow("INSERT INTO dairy(dairy_id) VALUES($1) returning id;", dairy.DairyId).Scan(id);
	return err
}

func FindDairy(dairyId string, db *DB) (*Dairy) {

	var dairy *Dairy
	result := db.QueryRow("select * from dairy where dairy_id = $1", dairyId)
	var id int64
	err := result.Scan(&id)
	if err != nil {
		log.Fatalln("Error while retrieving dairy details", err)
		return dairy
	}

	dairy = &Dairy{id, dairyId}
	return dairy
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
