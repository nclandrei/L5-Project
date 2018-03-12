package db

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/nclandrei/L5-Project/jira"
	"gopkg.in/mgo.v2"
)

// MgoSession defines the connection to the MongoDB
type MgoSession struct {
	*mgo.Session
	dbName   string
	collName string
}

// NewDatabase returns a pointer to a Mongo Session
func NewDatabase(url, dbName, collName string) (*MgoSession, error) {
	session, err := mgo.DialWithTimeout(url, 30*time.Second)
	session.SetMode(mgo.Monotonic, true)
	return &MgoSession{
		Session:  session,
		dbName:   dbName,
		collName: collName,
	}, err
}

// InsertIssues inserts a given slice of issues inside the default collection (i.e. issues)
func (db *MgoSession) InsertIssues(issues []jira.Issue) error {
	sessCopy := db.Copy()
	defer sessCopy.Close()
	errs := ""
	c := sessCopy.DB("nclandrei").C("issues")
	for _, issue := range issues {
		if err := c.Insert(issue); err != nil {
			errs += fmt.Sprintf("could not insert issue [%s]: %v\n", issue.Key, err)
		}
	}
	return fmt.Errorf(errs)
}

// GetIssues returns a collection of issues from the database
func (db *MgoSession) GetIssues(query bson.M) ([]jira.Issue, error) {
	var result []jira.Issue
	c := db.DB(db.dbName).C(db.collName)
	if err := c.Find(query).All(result); err != nil {
		return nil, err
	}
	return result, nil
}
