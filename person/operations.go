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

func HandlerCreatePerson(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var person Person
	err := parse(r, &person)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error while parsing json", err)))
		return
	}

	person.LastUpdated = time.Now()
	id, err := insertPerson(person, getDataBaseConnection())

	if err != nil {
		log.Println("error is", err)
		w.Write([]byte(fmt.Sprint("Error while writing to data base", err)))
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(person)
	log.Println("Inserted data base id", id)
}
func parse(r *http.Request, dataType interface{}) (error) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(dataType)
	return err
}

func insertPerson(person Person, db *DB) (int64, error) {
	log.Println("Person deatils", person)
	var id int64
	err := db.QueryRow("INSERT INTO persons(person_id,first_name,last_name,last_updated) VALUES($1,$2,$3, $4) returning id;",
		person.PersonId, person.FirstName, person.LastName, person.LastUpdated).Scan(&id)
	return id, err
}

func getDataBaseConnection() *DB {
	db, err := Open("postgres", "host=localhost port=1111 dbname=kmfdetails user=kmfadmin password=changeme001 sslmode=disable")
	if err != nil {
		panic("Error while getting the data base connection")
	}
	return db
}
