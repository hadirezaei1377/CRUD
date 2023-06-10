package main

import (
	"CRUD/logger"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// task :
// use DB , instead save to file use some functions for connicting to database(sqlite), dont use gorm,
// sqlite in a package and postgress in another package
// second functions like DeleteRecord be in a interface
// handlers in a seperated file

// how can I improve that ?
// 1- Add input validation
// 2- Use a database
// 3- Logging
// 4- Authentication and authorization

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

	dbType := "postgres"

	var db *sql.DB
	var err error

	switch dbType {
	case "postgres":
		connStr := "host=your_host port=your_port user=your_user password=your_password dbname=your_dbname sslmode=require"
		db, err = sql.Open("postgres", connStr)
	case "sqlite":
		db, err = sql.Open("sqlite3", "./foo.db")
	default:
		log.Fatalf("Unknown database type: %s", dbType)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()

	Migrate(db)

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
		r.HandleFunc(route.Path, director(db)).Methods(route.Method)
	}
	r.HandleFunc("/records", func(w http.ResponseWriter, r *http.Request) {
		GetRecords(w, r, db)
	}).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func director(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetRecords(w, r, db)
		case http.MethodPost:
			AddRecord(w, r, db)
		case http.MethodPut:
			UpdateRecordByID(w, r, db)
		case http.MethodDelete:
			DeleteRecordByID(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func DeleteRecordByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	panic("unimplemented")
}

func UpdateRecordByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	err = UpdateRecord(id, &article, db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateRecord(id int, article *Article, db *sql.DB) error {
	var existingArticle Article
	err := db.QueryRow("SELECT id, title, description, created_date FROM articles WHERE id=$1", id).Scan(&existingArticle.ID, &existingArticle.Title, &existingArticle.Description, &existingArticle.CreatedDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("Article with ID %d not found", id)
		}
		return err
	}

	existingArticle.Title = article.Title
	existingArticle.Description = article.Description

	_, err = db.Exec("UPDATE articles SET title=$1, description=$2 WHERE id=$3", existingArticle.Title, existingArticle.Description, id)
	if err != nil {
		return fmt.Errorf("Failed to update article: %v", err)
	}

	return nil
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

func GetRecordByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var record Article
	err := db.QueryRow("SELECT id, title, description, created_date FROM articles WHERE id=$1", id).Scan(&record.ID, &record.Title, &record.Description, &record.CreatedDate)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(record)
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			created_date TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)

	if err != nil {
		return fmt.Errorf("Failed to migrate database: %v", err)
	}

	return nil
}

func ShowData(db *sql.DB) (dataStore []Article, err error) {
	rows, err := db.Query("SELECT id, title, description, created_date FROM articles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var article Article
		err := rows.Scan(&article.ID, &article.Title, &article.Description, &article.CreatedDate)
		if err != nil {
			return nil, err
		}
		dataStore = append(dataStore, article)
	}

	return dataStore, nil
}

func GetRecords(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	dataStore, err := ShowData(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(dataStore)
}

func AddRecord(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var newArticle Article
	err := json.NewDecoder(r.Body).Decode(&newArticle)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = db.QueryRow("INSERT INTO articles (title, description, created_date) VALUES ($1, $2, NOW()) RETURNING id", newArticle.Title, newArticle.Description).Scan(&newArticle.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add record: %v", err), http.StatusInternalServerError)
		return
	}

	response := Response{Msg: "Article added successfully"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
