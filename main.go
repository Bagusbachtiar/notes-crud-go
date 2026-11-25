package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dbPool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	store := &NoteStore{db: dbPool}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /notes", func(w http.ResponseWriter, r *http.Request){
		var input Note
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
			http.Error(w, "Title and content are required", http.StatusBadRequest)
			return
		}

		if err := store.CreateNote(r.Context(), input.Title, input.Content); err != nil{
			log.Printf("DB Insert Error: %v", err)
			http.Error(w, "Failed to create note", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Note Created!"})
	})

	mux.HandleFunc("GET /notes", func(w http.ResponseWriter, r *http.Request){
		notes, err := store.GetAllNotes(r.Context())
		if err != nil {
			log.Printf("DB Query Error: %v", err)
			http.Error(w, "Failed to fetch Notes", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notes)
	})

	log.Println("📝 Notes CRUD API running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}