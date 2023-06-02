package main

import (
	"CRUD/logger"
	"database/sql"
	"encoding/json"
	"fmt"
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
	r.Methods("GET").HandleFunc("/records/{id}", GetRecordByID)
	r.Methods(http.MethodPut).HandleFunc("/records/{id}", UpdateRecordByID)
	r.Methods(http.MethodDelete).HandleFunc("/records/{id}", DeleteRecordByID)

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

func DeleteRecordByID(w http.ResponseWriter, r *http.Request) {
	// extract id from the URL params
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = DeleteRecord(id)
	if err != nil {
		// handle error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

func ShowData() (dataStore []Article) { // display or manipulate data    // todo : line by line
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

	// DB

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert the new record into the articles table
	query := `INSERT INTO articles (title, description, created_date) VALUES ($1, $2, $3)`
	_, err = db.Exec(query, newArticle.Title, newArticle.Description, newArticle.CreatedDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Response{Msg: "Article added successfully"}
	json.NewEncoder(w).Encode(response)

}

// connecting to postgressql    // todo : line by line
func ConnectToDB() (*sql.DB, error) {

	dbHost := "localhost"
	dbPort := "5432"
	dbUser := "username"
	dbPassword := "password"
	dbName := "dbname"

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	fmt.Println("you are connected to database successfully")

	return db, nil
}
