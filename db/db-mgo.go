package db

import (
	"github.com/nclandrei/L5-Project/jira"
	"gopkg.in/mgo.v2"
	"time"
)

// MgoSession defines the connection to the MongoDB
type MgoSession struct {
	*mgo.Session
	dbName   string
	collName string
}

// NewDatabase returns a pointer to a Mongo Session
func NewDatabase(url, dbName, collName string) (*MgoSession, error) {
	session, err := mgo.DialWithTimeout(url, 60*time.Second)
	session.SetPoolLimit(5000)
	session.SetMode(mgo.Monotonic, true)
	return &MgoSession{
		Session:  session,
		dbName:   dbName,
		collName: collName,
	}, err
}

// InsertIssue inserts a given slice of issues inside the default collection (i.e. issues)
func (db *MgoSession) InsertIssue(issue jira.Issue) error {
	db.Refresh()
	sessCopy := db.Session.Copy()
	defer sessCopy.Close()
	c := sessCopy.DB(db.dbName).C(db.collName)
	return c.Insert(issue)
}

// GetIssues returns a collection of issues from the database
func (db *MgoSession) GetIssues(query string) ([]byte, error) {
	var result []byte
	c := db.DB(db.dbName).C(db.collName)
	if err := c.Find(query).All(result); err != nil {
		return nil, err
	}
	return result, nil
}
