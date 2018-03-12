package db

import (
	"log"
	"time"

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
func InsertIssue(errChan chan error, iChan chan []jira.Issue, session *mgo.Session) {
	defer session.Close()
	issues := <-iChan
	log.Println("got inside db")
	c := session.DB("nclandrei").C("issues")
	for _, issue := range issues {
		if err := c.Insert(issue); err != nil {
			errChan <- err
		}
	}
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
