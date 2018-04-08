package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nclandrei/L5-Project/jira"

	"github.com/boltdb/bolt"
)

// Name of the bucket where we'll be inserting our users.
const (
	bucketName = "users"
)

// Database defines a generic interface for different DBs to implement.
type Database interface {
	Issues() ([]jira.Issue, error)
	InsertIssues([]jira.Issue) error
}

// BoltDB holds the information related to an instance of Bolt Database.
type BoltDB struct {
	*bolt.DB
}

// NewBoltDB returns a new Bolt Database instance.
func NewBoltDB(path string) (*BoltDB, error) {
	options := &bolt.Options{
		Timeout: 20 * time.Second,
	}
	db, err := bolt.Open(path, 0600, options)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, txErr := tx.CreateBucketIfNotExists([]byte(bucketName))
		err = txErr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &BoltDB{
		DB: db,
	}, err
}

// InsertIssues takes a slice of issues and inserts them into Bolt.
func (db *BoltDB) InsertIssues(issues ...jira.Issue) error {
	for _, issue := range issues {
		tx, err := db.Begin(true)
		if err != nil {
			return fmt.Errorf("could not create transaction: %v", err)
		}
		b := tx.Bucket([]byte(bucketName))
		buf, err := json.Marshal(&issue)
		if err != nil {
			return fmt.Errorf("could not marshal issue %s: %v", issue.Key, err)
		}
		err = b.Put([]byte(issue.Key), buf)
		if err != nil {
			return fmt.Errorf("could not insert issue %s: %v", issue.Key, err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("could not commit transaction: %v", err)
		}
	}
	return nil
}

// IssueByKey returns a single issue searched for by key.
func (db *BoltDB) IssueByKey(key string) (*jira.Issue, error) {
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return nil, fmt.Errorf("could not retrieve users bucket from bolt")
	}
	var issue *jira.Issue
	bIssue := b.Get([]byte(key))
	if bIssue == nil {
		return nil, nil
	}
	err = json.Unmarshal(bIssue, &issue)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

// Issues retrieves all the issues from inside the database.
func (db *BoltDB) Issues() ([]jira.Issue, error) {
	var issues []jira.Issue
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return nil, fmt.Errorf("could not retrieve users bucket from bolt")
	}
	err = b.ForEach(func(k, v []byte) error {
		var issue jira.Issue
		err := json.Unmarshal(v, &issue)
		if err == nil {
			issues = append(issues, issue)
		}
		return err
	})
	return issues, err
}

// Cursor returns a cursor to the users inside the bucket as well as a function to close the open tx.
func (db *BoltDB) Cursor() (*bolt.Cursor, func() error, error) {
	tx, err := db.Begin(true)
	if err != nil {
		return nil, nil, err
	}
	b := tx.Bucket([]byte(bucketName))
	teardown := func() error {
		return tx.Rollback()
	}
	return b.Cursor(), teardown, nil
}
