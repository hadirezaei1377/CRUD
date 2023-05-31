package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// show data
type Article struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"time"`
}
type Response struct {
	Msg string
}

func main() {

	ShowData() // show existing data from file

	http.HandleFunc("/records", getRecords)
	http.HandleFunc("/records/add", addRecord) // additiona posts

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getRecords(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ShowData())
}

func addRecord(w http.ResponseWriter, r *http.Request) {
	// First read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()

	// Then open the file for next reads and writes
	file, err := os.OpenFile("data.json", os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read previous file content for append
	file_content, _ := io.ReadAll(file)

	// Write all file content
	// TODO: This is wrong because we can't attach two json strings
	new_content := append(body, file_content...)
	// Previous content may exist and duplicate after every new post request
	_, err = file.Write(new_content)

	// Return error when can't save post
	if err != nil {
		resp, _ := json.Marshal(Response{
			Msg: "can't save post",
		})
		w.Header().Set("status", "400")
		w.Write(resp)
	}
}

// show data
func ShowData() (dataStore []Article) {
	file, err := os.OpenFile("data.json", os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	bytes, err := io.ReadAll(file) // store in binary mode
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &dataStore)
	if err != nil {
		log.Fatal(err)
	}
	return
}
