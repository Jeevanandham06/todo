package main

import (
	"log"
	"net/http"
	"todo/addtaa"
	"todo/show"
	"todo/updel"
)

func main() {
	// Routes
	http.HandleFunc("/add", addtaa.AddTask)
	http.HandleFunc("/show", show.ShowTask)
	http.HandleFunc("/upd", updel.UpdateTask)
	http.HandleFunc("/del", updel.DeleteTask) // Now uses PUT method

	// Start server
	port := ":8082"
	log.Printf("Server running at http://localhost%s\n", port)
	log.Println("POST /add   - Add task (unique taskname required)")
	log.Println("GET /show   - Show all active tasks")
	log.Println("PUT /upd    - Update task details")
	log.Println("PUT /del    - Delete task (soft delete, status=N)")
	log.Println("Time format: HH:MM:SS (synced with local timezone)")
	log.Fatal(http.ListenAndServe(port, nil))
}
