package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// Initialize database
func initDB() {
	var err error

	db, err = sql.Open("sqlite", "history.db")
	if err != nil {
		log.Fatal("DB open error:", err)
	}

	// Create table if not exists
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		value TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("Table creation error:", err)
	}
}

// Handle incoming POST requests
func historyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var received []string

	err := json.NewDecoder(r.Body).Decode(&received)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(received) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("No data received"))
		return
	}

	fmt.Println("\n--- New History Batch ---")

	// Insert each value
	for _, v := range received {
		fmt.Println(v)

		_, err := db.Exec(
			"INSERT INTO history(value, created_at) VALUES(?, ?)",
			v,
			time.Now(),
		)
		if err != nil {
			log.Println("Insert error:", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Saved successfully"))
}

func main() {
	initDB()

	http.HandleFunc("/endpoint", historyHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

