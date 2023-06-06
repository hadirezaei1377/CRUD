package main

import (
	"CRUD/logger"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedDate time.Time `json:"created_date"`
}

type Response struct {
	Msg string
}

var globalLogger *zap.Logger

func main() {

	globalLogger = logger.InitializeLogger()

	Migrate() // data.json is existed or not , if not craete that

	articles = ShowData()

	r := mux.NewRouter()
	r.HandleFunc("/records", director)
	r.HandleFunc("/records/{id}", GetRecordByID)
	r.HandleFunc("/records/{id}", UpdateRecordByID)
	r.HandleFunc("/records/{id}", DeleteRecordByID)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func director(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Handle GET request to /records
		GetRecords(w, r)

	case http.MethodPost:
		AddRecord(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func UpdateRecordByID(w http.ResponseWriter, r *http.Request) {
	// extract id from the URL params
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// decode request body into an Article struct
	var article Article
	err = json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// update the record with the given id
	err = UpdateRecord(id, &article)
	if err != nil {
		// handle error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

var articles []Article

func UpdateRecord(id int, article *Article) error {
	// find the article with the given id in the slice
	for i, a := range articles {
		if a.ID == id {
			// update the fields of the article with the new values
			articles[i].Title = article.Title
			articles[i].Description = article.Description
			return nil
		}
	}

	// if no article was found with the given id, return an error
	return fmt.Errorf("Article with ID %d not found", id)
}

func DeleteRecordByID(w http.ResponseWriter, r *http.Request) {
	// extract id from the URL params
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := DeleteRecord(id); err != nil {
		// handle error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteRecord(id int) error {
	// Get all records from data store
	dataStore := ShowData()

	// Find the index of the record with the given ID
	index := -1
	for i, article := range dataStore {
		if article.ID == id {
			index = i
			break
		}
	}

	// Return an error if the record is not found
	if index == -1 {
		return fmt.Errorf("Record not found")
	}

	// Remove the record from the data store
	dataStore = append(dataStore[:index], dataStore[index+1:]...)

	// Save the updated data back to the data store file
	file, err := os.OpenFile("data.json", os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(dataStore)
	if err != nil {
		return err
	}

	return nil
}
func GetRecordByID(w http.ResponseWriter, r *http.Request) { // todo : db and line by line

	var record Article

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	dataStore := ShowData()

	for _, article := range dataStore {
		if article.ID == id {
			record = article
			break
		}
	}

	if record.ID == 0 { // Return an error response if the record is not found
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(record)
}

func Migrate() {
	file, err := os.OpenFile("data.json", os.O_CREATE, 0644)
	if err != nil {
		globalLogger.Error(err.Error())
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		globalLogger.Error(err.Error())
		return
	}

	if stat.Size() == 0 {
		file.Write([]byte("[]"))
	}

}

func ShowData() (dataStore []Article) {
	// Open data store file
	file, err := os.OpenFile("data.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read JSON data from file
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal JSON data into dataStore variable
	err = json.Unmarshal(bytes, &dataStore)
	if err != nil {
		log.Fatal(err)
	}

	return dataStore
}

func GetRecords(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ShowData())
}

func AddRecord(w http.ResponseWriter, r *http.Request) {

	var newArticle Article

	err := json.NewDecoder(r.Body).Decode(&newArticle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	newArticle.CreatedDate = time.Now()

	response := Response{Msg: "Article added successfully"}

	json.NewEncoder(w).Encode(response)

}
