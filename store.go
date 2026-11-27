package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

//Notestore holds our neon db Connection
type NoteStore struct {
	db *pgxpool.Pool
}

//CreateNote inserts a new row into neon postgress (like Note::create() in Laravel)
//* is a pointer to existing var NoteStore that already processed by RAM so when u have same Var construct to process u use * to save memmory. 
func (s *NoteStore) CreateNote(ctx context.Context, title, content string) error {
	query := `INSERT INTO notes (title, content) VALUES ($1, $2)`
	//_ throw 1st values which is succed and take 2nd values which is error. in EXEC only have/need to get 1 values which is error. no need succes return.
	_, err := s.db.Exec(ctx, query, title, content)
	return err
}

//GetAllNotes fetches every note from neon porstgres (like Note::all() in Laravel)
func (s *NoteStore) GetAllNotes(ctx context.Context) ([]Note, error) {
	query := `SELECT id, title, content, created_at FROM notes ORDER BY id DESC`

	//in Query return 2 value not like Exec. rows in the db, and the error
	rows, err := s.db.Query(ctx, query)
	// this means if err != (is not Equal) to nil (zero) if its error then its not nil (zero+) so run the {}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//empty list from struct model in model.go
	var notes []Note

	//for is loop, rows.Next is pointing the loop to get the next Row and parentheses is to trigger it wether u get the row or not
	for rows.Next() {
		//making new n var from struct Note in model.go
		var n Note
		// 1st part of this is catching error in rows.Scan after ; is to see if the err is nil (zero) if its not zero then return nil, err
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			return nil, err
		}
		//the n data that already scanned and correct is inserted to notes. which is Note struct in model.go
		notes = append(notes, n)
	}
	return notes, nil
}