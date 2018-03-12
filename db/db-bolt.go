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
	timeFormat = "2006-01-02 15:04:05 MST"
)

// BoltDB holds the information related to an instance of Bolt Database
type BoltDB struct {
	*bolt.DB
	path string
}

// NewBoltDB returns a new Bolt Database instance
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
		DB:   db,
		path: path,
	}, err
}

// InsertIssues takes a slice of issues and inserts them into Bolt
func (db *BoltDB) InsertIssues(issues []jira.Issue) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		for _, issue := range issues {
			buf, err := json.Marshal(&issue)
			if err != nil {
				return err
			}
			err = b.Put([]byte(issue.Key), buf)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetIssues retrieves all the issues from inside the database
func (db *BoltDB) GetIssues() ([]jira.Issue, error) {
	var issues []jira.Issue
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("could not retrieve users bucket from bolt")
		}
		return b.ForEach(func(k, v []byte) error {
			var issue jira.Issue
			err := json.Unmarshal(v, &issue)
			if err == nil {
				issues = append(issues, issue)
			}
			return err
		})
	})
	return issues, err
}
