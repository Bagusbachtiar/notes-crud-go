package main

import (
	//gives context.Background() so can connect to postgress
	"context"
	// translate incoming Json into GO Structs (NewDecoder) and GO into Json (NewEncoder)
	"encoding/json"
	//log prints timestamps and message into terminal (log.Println, log.Fatalf)
	"log"
	//Go built in web server and router (http.NewServerMux, http.ListenAndServe, http.Error)
	"net/http"
	//Reads computer environtment variables (os.Getemv("DATABASE_URL"))
	"os"
	//postgress driver that creates dbPool Connection
	"github.com/jackc/pgx/v5/pgxpool"
	//Loads key value pairs from .env file into memmory
	"github.com/joho/godotenv"
)

func main() {
	//godotenv reads local .env file from the start of the app
	//_ = is to ignore Error when app is deplyoed in cloud server cause when its deploye env file is not in the main code folder its in cloud dashboard godotenv.
	// -Load() returns a "file not found" error. Using _ = ignores that error so your app doesn't crash in production.
	_ = godotenv.Load()
	
	//this brand new variable means i wanted 2 value from this process which is dbconnect and the error if failed
	//and to get a new connection to db u use this (pgxpool.New) and (make a space for new running connecting program using context.Background() and if u already make the space then u just call the db using os.Getenv("DATABASE_URL")) 
	dbPool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	//read if err is not equal to nil (zero) = means error
	if err != nil {
		//log.Fatalf is stopping the DB when fails to connect
		//%v is a actual connection error and ",err" is the error value
		log.Fatalf("failed to connect to database: %v", err)
	}
	//this just closing the connection when not needed. BUT if the code above returns err then it wont even need this.cause Fatalf just terminates the entire program
	defer dbPool.Close()

	//new var store is a db worker which NoteStore is defined at the top of store.go along with its work which is storing data
	//{db: dbPool} is called (struct literal, { fieldName: value }) it tells NoteStore that are in store.go to Input the data into connected db 
	store := &NoteStore{db: dbPool}
	
	//new var called mux. taken from  Go's classic technical term for an HTTP router/switchboard "Multiplexer"
	//NewServeMux is a constructor from Go's Standard "net/http" package to create a brand new empty routing table
	//net/http is like the path net is the folder http the actual package and then () parentheses to execute func
	mux := http.NewServeMux() 

	//mux.HandleFunc is a method, mux is from var above and HanldeFunc is a method attached to mux
	//mux.HandleFunc takes two inputs inside its parentheses: the URL pattern ("POST /notes") and an anonymous function (func(w http.ResponseWriter, r *http.Request)) that runs whenever a user hits that URL
	//1. The Route, POST /notes . this tells mux to trigger this function when http post req arrives at notes like in Laravel route::post('/notes', ...)
	//2. Inline Handler inside inline Handler is interface (w http.ResponseWriter) and a struct taken from net/http (r http.Request)
	// w http.ResponseWriter, w is a var and http.ResponseWriter is a outgoing response headers status codes (201, 400, 500). so its to send messsage if its err or not.
	// r http.Request, r is also a var and http.Request is a struct handlingincoming Request
	// * is a pointer to existing var http.Request that already processed by RAM so when u have same Var construct to process u use * to save memmory. 
	mux.HandleFunc("POST /notes", func(w http.ResponseWriter, r *http.Request){
	
		//new var, input is a variable also but for empty box like in Laravel ($store/$update)
		//Note is a defined struct blueprint inside store.go like model in Laravel (ID, Title, Content, CreatedAt)
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