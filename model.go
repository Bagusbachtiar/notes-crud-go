package main

import "time"

//Note is our Model (just like Note.php in Laravel).
// Each field matches one column in your Neon Postgres 'notes' table.
type Note struct {
	ID			int			`json:"id"`
	Title 		string		`json:"title"`
	Content		string		`json:"content"`
	CreatedAt	time.Time	`json:"created_at"`
}