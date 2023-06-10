package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"CRUD/Article"
)

// initialize the PostgreSQL database connection
func Connect() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("PGHOST"), 5432, os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func SaveData(db *sql.DB, dataStore []Article) error {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO articles (data) VALUES ($1)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, article := range dataStore {
		data, err := json.Marshal(article)
		if err != nil {
			return err
		}

		_, err = stmt.ExecContext(ctx, data)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func ShowData(db *sql.DB) ([]Article, error) {
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, `SELECT data FROM articles`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataStore []Article
	for rows.Next() {
		var data []byte
		err := rows.Scan(&data)
		if err != nil {
			return nil, err
		}

		Article := 
		var article Article
		err = json.Unmarshal(data, &article)
		if err != nil {
			return nil, err
		}

		dataStore = append(dataStore, article)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return dataStore, nil
}
