package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	err = Migrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

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

func AddArticle(ds *Database) func(http.ResponseWriter, *http.Request) {
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

func GetArticleByID(ds *Database) func(http.ResponseWriter, *http.Request) {
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

func UpdateArticleByID(ds *Database) func(http.ResponseWriter, *http.Request) {
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
			http.Error(w, fmt.Sprintf("Failed to update record: %v", err), http.StatusInternalServerError)
			return
		}

		response := Response{Msg: "Article updated successfully"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func DeleteArticleByID(ds *Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, _ := strconv.Atoi(vars["id"])

		err := ds.DeleteRecord(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete record: %v", err), http.StatusInternalServerError)
			return
		}

		response := Response{Msg: "Article deleted successfully"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

}

func SaveData(ds *Database) error {
	articles, err := ds.ShowData()
	if err != nil {
		return err
	}

	for _, article := range articles {
		err = ds.AddRecord(&article)
		if err != nil {
			return err
		}
	}

	return nil
}

func Migrate(db *sql.DB) error {
	query := "CREATE TABLE IF NOT EXISTS articles (id SERIAL PRIMARY KEY, title TEXT, description TEXT, created_date TIMESTAMP);"
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (ds *Database) ShowData() ([]Article, error) {
	articles := []Article{}
	rows, err := ds.db.Query("SELECT id, title, description, created_date FROM articles ORDER BY created_date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a Article
		err = rows.Scan(&a.ID, &a.Title, &a.Description, &a.CreatedDate)
		if err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func (ds *Database) AddRecord(a *Article) error {
	query := "INSERT INTO articles (title, description, created_date) VALUES ($1, $2, $3) RETURNING id"
	err := ds.db.QueryRow(query, a.Title, a.Description, time.Now()).Scan(&a.ID)
	if err != nil {
		return err
	}

	return nil
}

func (ds *Database) GetRecord(id int, a *Article) error {
	query := "SELECT id, title, description, created_date FROM articles WHERE id = $1"
	err := ds.db.QueryRow(query, id).Scan(&a.ID, &a.Title, &a.Description, &a.CreatedDate)
	if err != nil {
		return err
	}

	return nil
}

func (ds *Database) UpdateRecord(id int, a *Article) error {
	query := "UPDATE articles SET title=$2, description=$3 WHERE id=$1"

	_, err := ds.db.Exec(query, id, a.Title, a.Description)
	if err != nil {
		return err
	}

	return nil
}

func (ds *Database) DeleteRecord(id int) error {
	query := "DELETE FROM articles WHERE id=$1"

	_, err := ds.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func GetArticles(ds *Database) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		articles, err := ds.ShowData()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to retrieve records: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(articles)
	}
}
