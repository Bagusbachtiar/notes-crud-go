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
	// r http.Request, r is also a var and http.Request is a struct handling incoming Request
	// * is a pointer to existing var http.Request that already processed by RAM so when u have same Var construct to process u use * to save memmory. 
	mux.HandleFunc("POST /notes", func(w http.ResponseWriter, r *http.Request){
	
		//new var, input is a variable also but for empty box like in Laravel ($store/$update)
		//Note is a defined struct blueprint inside store.go like model in Laravel (ID, Title, Content, CreatedAt)
		var input Note
		
		//json.Decoder is to create Json reader pointing to the raw Json body
		//Decode is to reads the Json text and translate it into Go Fields like `json:"id"` in the model.go and input the values directly
		//& is to decode input directly to Note box so it fills struct directly with out creating duplicates in ram
		//After ; || means OR, so basically if err OR input Title isnt there 
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
			//http.Error is a helper function from net/http
			//w is the var for http.ResponseWriter so http.Error knows where to send reply
			//http.StatusBadRequest is a built in constan from net.http. which holds the HTTP status code 400
			http.Error(w, "Title and content are required", http.StatusBadRequest)
			return
		}

		//store is local Var use to connect the db, and  NoteStore ia a struct inside store.go to use pgxpool
		//CreateNote is a methode also defined in store.go
		//input is var for a empty box like $store in laravel. then it calls Title calling the Var naming in model.go
		if err := store.CreateNote(r.Context(), input.Title, input.Content); err != nil{
			//this is to print the log message and %v is to copy the log and put it in terminal, then err is values or argument
			log.Printf("DB Insert Error: %v", err)
			//also built in constan from net/http. then throw the error which this one is from server  
			http.Error(w, "Failed to create note", http.StatusInternalServerError)
			return
		}

		// this 3 line is like [return response()->json(['message' => 'Note Created!'], 201);] in Laravel
		//Sending json format to FE 
		w.Header().Set("Content-Type", "application/json")
		//Sends the HTTP status code 201 Created or a succes code 
		w.WriteHeader(http.StatusCreated)
		//is translating GO backend language to Json to send to FE 
		json.NewEncoder(w).Encode(map[string]string{"message": "Note Created!"})
	})

	//mux is a multiplexer. technical term for an HTTP router
	//HandleFunc is method and has 2 input in parentheses, the URL pattern and the inline handler. 
	// for the inline handler here is the ResponseWriter and the Request
	mux.HandleFunc("GET /notes", func(w http.ResponseWriter, r *http.Request){
		//notes is a new local var 
		//store local var and its a worker that are defined in store.go Named NoteStore 
		//GetAllNotes is a methode function defined in store.go which is getting data from DB using the user's request context (r.Context()). 
		notes, err := store.GetAllNotes(r.Context())
		if err != nil {
			//same error handler 
			log.Printf("DB Query Error: %v", err)
			http.Error(w, "Failed to fetch Notes", http.StatusInternalServerError)
			return
		}

		//Changing Go and using JSON to send Succes message and also the data to Json
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notes)
	})

	//show message in terminal if Program is running or error. when error then just terminates program
	log.Println("📝 Notes CRUD API running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}