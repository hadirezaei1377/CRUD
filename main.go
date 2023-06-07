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
	Msg string `json:"msg"`
}

var globalLogger *zap.Logger
var articles []Article

func main() {
	globalLogger = logger.InitializeLogger()

	Migrate()

	articles = ShowData()

	r := mux.NewRouter()

	routes := []struct {
		Path   string
		Method string
	}{
		{"/records", http.MethodGet},
		{"/records", http.MethodPost},
		{"/records/{id}", http.MethodGet},
		{"/records/{id}", http.MethodPut},
		{"/records/{id}", http.MethodDelete},
	}

	for _, route := range routes {
		r.HandleFunc(route.Path, director).Methods(route.Method)
	}

	log.Fatal(http.ListenAndServe(":8080", r))
}

func director(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetRecords(w, r)
	case http.MethodPost:
		AddRecord(w, r)
	case http.MethodPut:
		UpdateRecordByID(w, r)
	case http.MethodDelete:
		DeleteRecordByID(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func UpdateRecordByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var article Article
	err = json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = UpdateRecord(id, &article)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateRecord(id int, article *Article) error {
	for i := range articles {
		if articles[i].ID == id {

			articles[i].Title = article.Title
			articles[i].Description = article.Description

			err := SaveData(articles)
			if err != nil {
				return fmt.Errorf("failed to save data: %v", err)
			}

			return nil
		}
	}

	return fmt.Errorf("Article with ID %d not found", id)
}

func SaveData(dataStore []Article) error {
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

func DeleteRecordByID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := DeleteRecord(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteRecord(id int) error {

	dataStore := ShowData()

	index := -1
	for i, article := range dataStore {
		if article.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("Record not found")
	}

	dataStore = append(dataStore[:index], dataStore[index+1:]...)

	err := SaveData(dataStore)
	if err != nil {
		return fmt.Errorf("Failed to save data: %v", err)
	}

	return nil
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

	if record.ID == 0 {
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

	file, err := os.OpenFile("data.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newArticle.ID = len(articles) + 1

	newArticle.CreatedDate = time.Now()

	articles = append(articles, newArticle)

	err = SaveData(articles)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Response{Msg: "Article added successfully"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
