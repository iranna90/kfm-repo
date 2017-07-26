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
)

type Person struct {
	id          int64
	PersonId    string `json:"personId"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	LastUpdated time.Time `json:"lastUpdated"`
}

var dataBaseConnection = getDataBaseConnection

func HandleGetPerson(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	personId := params["personId"]
	person := findPerson(personId, dataBaseConnection())

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
	defer r.Body.Close()
	var person Person

	err := decode(r, &person)
	if err != nil {
		log.Println("Error while parsing payload to person", err)
		http.Error(w, fmt.Sprint("Error while parsing payload to person", err), http.StatusBadRequest)
		return
	}

	db := dataBaseConnection()

	retrievedPerson := findPerson(person.PersonId, db)

	person.LastUpdated = time.Now()

	// check person exists
	fmt.Println("retrived person", retrievedPerson)
	if retrievedPerson != nil {
		// update
		log.Println("Person ", person.PersonId, "already exists,So updating it")
		err = updatePerson(&person, db)
	} else {
		// insert
		log.Println("Person ", person.PersonId, "does not exists,So inserting it")
		_, err = insertPerson(&person, db)
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

func getDataBaseConnection() *DB {
	db, err := Open("postgres", "host=localhost port=1111 dbname=kmfdetails user=kmfadmin password=changeme001 sslmode=disable")
	if err != nil {
		panic("Error while getting the data base connection")
	}
	return db
}

func findPerson(personId string, db *DB) (*Person) {

	var person *Person
	result := db.QueryRow("select * from persons where person_id = $1", personId)
	var (
		id                               int64
		person_id, first_name, last_name string
		last_update                      time.Time
	)
	err := result.Scan(&id, &person_id, &first_name, &last_name, &last_update)
	if err != nil {
		return person
	}
	person = &Person{id, person_id, first_name, last_name, last_update}

	return person
}

func insertPerson(person *Person, db *DB) (int64, error) {
	var id int64
	err := db.QueryRow("INSERT INTO persons(person_id,first_name,last_name,last_updated) VALUES($1,$2,$3, $4) returning id;",
		person.PersonId, person.FirstName, person.LastName, person.LastUpdated).Scan(&id)
	return id, err
}

func updatePerson(person *Person, db *DB) (err error) {
	query := "UPDATE persons SET first_name = $1, last_name = $2, last_updated = $3 WHERE person_id =$4"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println("Error while creating statement to update person", person.PersonId, err)
		return
	}
	_, err = stmt.Exec(person.FirstName, person.LastName, person.LastUpdated, person.PersonId)

	return
}
