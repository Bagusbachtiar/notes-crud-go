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
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//Create ana empty slice (list) to hold notes we read from DB 
	var notes []Note

	for rows.Next() {
		var n Note
		//copy the 4 columns from the current DB row into notes struct
		id err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}