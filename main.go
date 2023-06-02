package main

import (
	"CRUD/logger"
	"encoding/json"
	"io"
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

	r := mux.NewRouter()
	r.HandleFunc("/records", director)
	r.HandleFunc("/records/{id}", GetRecordByID).Methods("GET")
	r.HandleFunc("/records/{id}", UpdateRecordByID).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func director(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/records" { // Handle GET request to /records
			GetRecords(w, r)
		} else { // Handle GET request to /records/{id}
			GetRecordByID(w, r)
		}
	case http.MethodPost:
		AddRecord(w, r)
	case http.MethodDelete:
		DeleteRecordByID(w, r)
	case http.MethodPut:
		UpdateRecordByID(w, r)
	}
}

func UpdateRecordByID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	dataStore := ShowData()

	found := false
	for i, article := range dataStore {
		if article.ID == id {
			found = true

			// Read and decode the updated record from the request body
			var updatedRecord Article
			err := json.NewDecoder(r.Body).Decode(&updatedRecord)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			// Update the fields of the record with the given ID
			dataStore[i].Title = updatedRecord.Title
			dataStore[i].Description = updatedRecord.Description

			break
		}
	}

	if !found { // Return an error response if the record is not found
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	jsonData, err := json.Marshal(dataStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := os.OpenFile("data.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Response{Msg: "Record updated successfully"}
	json.NewEncoder(w).Encode(response)
}

func DeleteRecordByID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	dataStore := ShowData()

	found := false
	for i, article := range dataStore {
		if article.ID == id {
			found = true

			// Remove the record with the given ID
			copy(dataStore[i:], dataStore[i+1:])
			dataStore = dataStore[:len(dataStore)-1]

			break
		}
	}

	if !found { // Return an error response if the record is not found
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	jsonData, err := json.Marshal(dataStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := os.OpenFile("data.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Response{Msg: "Record deleted successfully"}
	json.NewEncoder(w).Encode(response)

}

func GetRecordByID(w http.ResponseWriter, r *http.Request) {

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

func ShowData() (dataStore []Article) { // display or manipulate data
	file, err := os.OpenFile("data.json", os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		globalLogger.Error(err.Error())

	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		globalLogger.Error(err.Error())
	}

	err = json.Unmarshal(bytes, &dataStore) // datastore is a slice of articles
	if err != nil {
		globalLogger.Error(err.Error())
	}
	return
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

	dataStore := ShowData()
	newArticle.ID = len(dataStore) + 1

	// Add the new article to the end of the slice
	dataStore = append(dataStore, newArticle)

	jsonData, err := json.Marshal(dataStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := os.OpenFile("data.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Response{Msg: "Article added successfully"}
	json.NewEncoder(w).Encode(response)

}
