package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nclandrei/L5-Project/jira"

	"github.com/boltdb/bolt"
)

const (
	bucketName = "users"
)

// BoltDB holds the information related to an instance of Bolt Database.
type BoltDB struct {
	*bolt.DB
}

// NewBoltDB returns a new Bolt Database instance.
func NewBoltDB(path string) (*BoltDB, error) {
	options := &bolt.Options{
		Timeout: 30 * time.Second,
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
func (db *BoltDB) InsertIssues(issueChan chan []jira.Issue, errChan chan error) {
	for issues := range issueChan {
		tx, err := db.Begin(true)
		if err != nil {
			errChan <- fmt.Errorf("could not create transaction: %v", err)
		}
		b := tx.Bucket([]byte(bucketName))
		for _, issue := range issues {
			buf, err := json.Marshal(&issue)
			if err != nil {
				errChan <- fmt.Errorf("could not marshal issue %s: %v", issue.Key, err)
			}
			err = b.Put([]byte(issue.Key), buf)
			if err != nil {
				errChan <- fmt.Errorf("could not insert issue %s: %v", issue.Key, err)
			}
		}
		if err = tx.Commit(); err != nil {
			errChan <- fmt.Errorf("could not commit transaction: %v", err)
		}
	}
	close(errChan)
}

// GetAllIssues retrieves all the issues from inside the database.
func (db *BoltDB) GetAllIssues() ([]jira.Issue, error) {
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
