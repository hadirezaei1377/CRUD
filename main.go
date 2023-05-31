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

type Article struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedDate time.Time `json:"time"`
}

type Response struct {
	Msg string
}


func ShowData() (dataStore []Article) { // display or manipulate data
	file, err := os.OpenFile("data.json", os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &dataStore)  // datastore is a slice of articles
	if err != nil {
		log.Fatal(err)
	}
	return
}


func GetRecords(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ShowData())
}


func AddRecord(w http.ResponseWriter, r *http.Request) {

    var newArticle Article
    

	// First read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()


    

    dataStore := ShowData() // retrieve existing data store

    // Add the new article to the end of the slice
    dataStore = append(dataStore, newArticle)

    // Encode the updated data store as JSON and write it back to the file
	
	// Then open the file for next reads and writes
    file, err := os.OpenFile("data.json", os.O_WRONLY|os.O_TRUNC, 0644)    
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    bytes, err := json.MarshalIndent(dataStore, "", "  ")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    _, err = file.Write(bytes)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    response := Response{Msg: "Article added successfully"}
    json.NewEncoder(w).Encode(response)
}




// for json adding : read file and save it in an array , unmarshal apped new article and then marshal



 // file_content, _ := io.ReadAll(file)


 // new_content := append(body, file_content...)

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



func main() {

	ShowData()

	http.HandleFunc("/records", GetRecords)
	http.HandleFunc("/records/add", AddRecord) // additiona posts

	log.Fatal(http.ListenAndServe(":8080", nil))
}


