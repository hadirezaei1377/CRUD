package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedDate time.Time `json:"created_date"`
}

type Database struct {
	db *sql.DB
}

type Response struct {
	Msg string `json:"msg"`
}

var articles []Article

func main() {

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

	dataStore := &Database{db: db}

	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/articles", GetArticles(dataStore)).Methods(http.MethodGet)
	r.HandleFunc("/articles", AddArticle(dataStore)).Methods(http.MethodPost)
	r.HandleFunc("/articles/{id}", GetArticleByID(dataStore)).Methods(http.MethodGet)
	r.HandleFunc("/articles/{id}", UpdateArticleByID(dataStore)).Methods(http.MethodPut)
	r.HandleFunc("/articles/{id}", DeleteArticleByID(dataStore)).Methods(http.MethodDelete)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", r))
}

func GetArticles(ds dataStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		articles, err := ds.GetRecords()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(articles)
	}
}

func AddArticle(ds dataStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newArticle Article
		err := json.NewDecoder(r.Body).Decode(&newArticle)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err = ds.AddRecord(&newArticle)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to add record: %v", err), http.StatusInternalServerError)
			return
		}

		response := Response{Msg: "Article added successfully"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func GetArticleByID(ds dataStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, _ := strconv.Atoi(vars["id"])

		var article Article
		err := ds.GetRecord(id, &article)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Record not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(article)
	}
}

func UpdateArticleByID(ds dataStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		err = ds.UpdateRecord(id, &article)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func DeleteArticleByID(ds dataStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		err = ds.DeleteArticle(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := Response{Msg: "Article deleted successfully"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func (d *Database) DeleteArticle(id int) error {
	result, err := d.db.Exec("DELETE FROM articles WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("Failed to delete article: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func UpdateRecordByID(ds dataStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		err = ds.UpdateRecord(id, &article)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (d *Database) UpdateRecord(id int, article *Article) error {
	var existingArticle Article
	err := d.db.QueryRow("SELECT id, title, description, created_date FROM articles WHERE id=$1", id).Scan(&existingArticle.ID, &existingArticle.Title, &existingArticle.Description, &existingArticle.CreatedDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("Article with ID %d not found", id)
		}
		return err
	}

	existingArticle.Title = article.Title
	existingArticle.Description = article.Description

	_, err = d.db.Exec("UPDATE articles SET title=$1, description=$2 WHERE id=$3", existingArticle.Title, existingArticle.Description, id)
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

func GetRecordByID(w http.ResponseWriter, r *http.Request) {
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
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS articles (id SERIAL PRIMARY KEY, title TEXT NOT NULL, description TEXT NOT NULL, created_date TIMESTAMP NOT NULL DEFAULT NOW());")

	if err != nil {
		return fmt.Errorf("Failed to migrate database: %v", err)
	}

	return nil
}

func (d *Database) ShowData() ([]Article, error) {
	rows, err := d.db.Query("SELECT id, title, description, created_date FROM articles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]Article, 0)
	for rows.Next() {
		var article Article
		err := rows.Scan(&article.ID, &article.Title, &article.Description, &article.CreatedDate)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func GetRecords(ds dataStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		articles, err := ds.ShowData()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(dataStore)
	}
}

func AddRecord(ds dataStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newArticle Article
		err := json.NewDecoder(r.Body).Decode(&newArticle)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err = ds.AddRecord(&newArticle)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to add record: %v", err), http.StatusInternalServerError)
			return
		}

		response := Response{Msg: "Article added successfully"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
