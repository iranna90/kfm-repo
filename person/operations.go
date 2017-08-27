package person

import (
	"time"
	"net/http"
	"encoding/json"
	"log"
	_ "github.com/lib/pq"
	. "database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"kmf-repo/database"
	"kmf-repo/balance"
	"kmf-repo/dairy"
)

type Person struct {
	Id          int64 `json:"-"`
	PersonId    string `json:"personId,omitempty"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
}

type PersonError struct {
	PersonId string
	Message  string
}

func (p PersonError) Error() string {
	return fmt.Sprintf("Error while doing operation for person : %s and error details are %s", p.PersonId, p.Message)
}

var dataBaseConnection = database.GetDataBaseConnection

func HandleGetPerson(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	personId := params["personId"]
	dairyId := params["dairyId"]

	db := dataBaseConnection()
	dairy := dairy.FindDairy(dairyId, db)
	if dairy == nil {
		http.Error(w, fmt.Sprintf("Dairy deos not exists"), http.StatusNotFound)
		return
	}

	person := FindPerson(dairy.Id, personId, db)
	if person == nil {
		http.Error(w, fmt.Sprintf("Person %s does not exists", personId), http.StatusNotFound)
		return
	}

	err := encode(w, person)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while writing person as json payload is %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

func HandlerCreateOrUpdatePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	personId := params["personId"]
	dairyId := params["dairyId"]
	defer r.Body.Close()
	var person Person

	err := decode(r, &person)
	if err != nil {
		log.Println("Error while parsing payload to person", err)
		http.Error(w, fmt.Sprint("Error while parsing payload to person", err), http.StatusBadRequest)
		return
	}

	db := dataBaseConnection()
	dairy := dairy.FindDairy(dairyId, db)

	if dairy == nil {
		log.Println("Please check dairy details are correct", err)
		http.Error(w, fmt.Sprint("Dairy does not exists", err), http.StatusBadRequest)
		return
	}

	retrievedPerson := FindPerson(dairy.Id, personId, db)

	person.LastUpdated = time.Now()

	// check person exists
	if retrievedPerson != nil {
		// update
		log.Println("Person ", personId, "already exists, So updating it")
		err = updatePerson(dairy.Id, personId, &person, db)
	} else {
		// insert
		log.Println("Person ", personId, "does not exists, So inserting it")
		_, err = insertPerson(dairy.Id, personId, &person, db)
	}

	if err != nil {
		log.Println("Error while inserting/updating person ", err)
		http.Error(w, fmt.Sprint("Error while writing to data base", err), http.StatusInternalServerError)
		return
	}

	err = encode(w, person)
	if err != nil {
		log.Println("Error while writing the payload", err)
		http.Error(w, fmt.Sprint("Error while writing the payload", err), http.StatusInternalServerError)
		return
	}

	log.Println("record stored successfully")
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

func FindPerson(dairyRef int64, personId string, db *DB) (*Person) {

	var person *Person
	result := db.QueryRow("select * from persons where dairy_ref = $1 and person_id = $2", dairyRef, personId)
	var (
		id                               int64
		person_id, first_name, last_name string
		last_update                      time.Time
	)
	err := result.Scan(&id, &dairyRef, &person_id, &first_name, &last_name, &last_update)
	if err != nil {
		log.Println("Person not found because ", err)
		return person
	}
	person = &Person{Id: id, FirstName: first_name, LastName: last_name, LastUpdated: last_update}

	return person
}

func insertPerson(dairyRef int64, personId string, person *Person, db *DB) (int64, error) {
	var id int64
	err := db.QueryRow("INSERT INTO persons(dairy_ref, person_id,first_name,last_name,last_updated) VALUES($1,$2,$3, $4, $5) returning id;",
		dairyRef, personId, person.FirstName, person.LastName, person.LastUpdated).Scan(&id)

	if err != nil {
		return id, PersonError{personId, fmt.Sprintf("Error while inserting person %s", err.Error())}
	}

	err = balance.CreateRecord(dairyRef, id)

	if err != nil {
		return id, PersonError{personId, fmt.Sprintf("Error while inserting initial balance record %s", err.Error())}
	}

	log.Println("Person and Initial balance for person are successfully created")
	return id, err
}

func updatePerson(dairyRef int64, personId string, person *Person, db *DB) (err error) {
	query := "UPDATE persons SET first_name = $1, last_name = $2, last_updated = $3 WHERE person_id =$4 and dairy_ref = $5"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println("Error while creating statement to update person", personId, err)
		return
	}
	_, err = stmt.Exec(person.FirstName, person.LastName, person.LastUpdated, personId, dairyRef)

	return
}
