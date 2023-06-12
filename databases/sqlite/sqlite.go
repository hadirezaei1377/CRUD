package sqlite

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"CRUD/Article"
// )

// // initialize the SQLite database connection
// func Connect() (*sql.DB, error) {
// 	db, err := sql.Open("sqlite3", "./data.db")
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = db.Ping()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return db, nil
// }

// func SaveData(db *sql.DB, dataStore []Article) error {
// 	data, err := json.Marshal(dataStore)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = db.Exec(`INSERT INTO articles (data) VALUES (?)`, data)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func ShowData(db *sql.DB) ([]Article, error) {
// 	rows, err := db.Query(`SELECT data FROM articles`)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var dataStore []Article
// 	for rows.Next() {
// 		var data []byte
// 		err := rows.Scan(&data)
// 		if err != nil {
// 			return nil, err
// 		}

// 		Article :=
// 		var article Article
// 		err = json.Unmarshal(data, &article)
// 		if err != nil {
// 			return nil, err
// 		}

// 		dataStore = append(dataStore, article)
// 	}

// 	err = rows.Err()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return dataStore, nil
// }
