package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

// NewIssueDatabase returns a new database for the retrieved Jira issues
func NewIssueDatabase() (*sql.DB, error) {
	connStr := "user=nclandrei password=nclandrei dbname=nclandrei sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func ReadFromDB() {
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	rows, err := db.Query("SELECT * FROM ISSUES;")
	if err != nil {
		log.Fatalf("Could not query database for issues: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id sql.NullInt64
		var summary sql.NullString
		var description sql.NullString
		var comments sql.NullString
		var key string
		err = rows.Scan(&id, &summary, &description, &comments, &key)
		fmt.Printf("%v | %v | %v | %v | %v\n", id, summary, description, comments, key)
	}
}
