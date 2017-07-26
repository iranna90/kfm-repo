package person

import (
	"time"
	"net/http"
	"encoding/json"
	"log"
	_ "github.com/lib/pq"
	. "database/sql"
	"fmt"
)

type Person struct {
	id          int64
	PersonId    string `json:"personId"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	LastUpdated time.Time `json:"lastUpdated"`
}

var dataBaseConnection = getDataBaseConnection

func HandlerCreateOrUpdatePerson(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var person Person
	err := decode(r, &person)
	if err != nil {
		log.Println("Error while parsing payload to person", err)
		http.Error(w, fmt.Sprint("Error while parsing payload to person", err), 400)
		return
	}

	db := dataBaseConnection()

	retrievedPerson, err := findPerson(person.PersonId, db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while reading the person", err), 500)
	}
	
	// check person exists
	if retrievedPerson.id != 0 {
		// update	
	}else {
		// insert
	}
	person.LastUpdated = time.Now()
	id, err := insertPerson(person, db)

	if err != nil {
		log.Println("Error while inserting person ", err)
		http.Error(w, fmt.Sprint("Error while writing to data base", err), 500)
		return
	}

	err = encode(w, person)
	if err != nil {
		log.Println("Error while writing the payload", err)
		http.Error(w, fmt.Sprint("Error while writing the payload", err), 500)
		return
	}

	log.Println("Inserted data base id", id)
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

func findPerson(personId string, db *DB) (*Person, error) {
	result := db.QueryRow("select * from persons where person_id = $1", personId)
	var (
		id int64
		person_id, first_name, last_name string
		last_update time.Time
	)
	err := result.Scan(&id, &person_id, &first_name, &last_name, &last_update)
	if err != nil {
		log.Println("Error while reading person with ", person_id)
		return nil, err
	}
	return &Person{id, person_id, first_name, last_name, last_update}, err
}


func insertPerson(person Person, db *DB) (int64, error) {
	log.Println("Person deatils", person)
	var id int64
	err := db.QueryRow("INSERT INTO persons(person_id,first_name,last_name,last_updated) VALUES($1,$2,$3, $4) returning id;",
		person.PersonId, person.FirstName, person.LastName, person.LastUpdated).Scan(&id)
	return id, err
}

func updatePerson()  {
	
}