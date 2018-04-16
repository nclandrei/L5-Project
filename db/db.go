package db

import (
	"encoding/json"
	"fmt"
	"github.com/nclandrei/ticketguru/jira"
	"time"

	"github.com/boltdb/bolt"
)

// Name of the bucket where we'll be inserting our users.
const (
	bucketName = "users"
)

// TicketStorage defines a generic interface for different DBs to implement.
type TicketStorage interface {
	Tickets() ([]jira.Ticket, error)
	Insert(...jira.Ticket) error
	Slice(int, int) ([]jira.Ticket, error)
	Size() (int, error)
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
		return txErr
	})
	if err != nil {
		return nil, err
	}
	return &BoltDB{
		DB: db,
	}, err
}

// Insert takes a slice of tickets and inserts them into Bolt.
func (db *BoltDB) Insert(tickets ...jira.Ticket) error {
	for _, ticket := range tickets {
		tx, err := db.Begin(true)
		if err != nil {
			return fmt.Errorf("could not create transaction: %v", err)
		}
		b := tx.Bucket([]byte(bucketName))
		buf, err := json.Marshal(&ticket)
		if err != nil {
			return fmt.Errorf("could not marshal ticket %s: %v", ticket.Key, err)
		}
		err = b.Put([]byte(ticket.Key), buf)
		if err != nil {
			return fmt.Errorf("could not insert ticket %s: %v", ticket.Key, err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("could not commit transaction: %v", err)
		}
	}
	return nil
}

// TicketByKey returns a single ticket searched for by key.
func (db *BoltDB) TicketByKey(key string) (*jira.Ticket, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return nil, fmt.Errorf("could not retrieve users bucket from bolt")
	}
	var ticket *jira.Ticket
	bTicket := b.Get([]byte(key))
	if bTicket == nil {
		return nil, nil
	}
	err = json.Unmarshal(bTicket, &ticket)
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

// Tickets retrieves all the tickets from inside the database.
func (db *BoltDB) Tickets() ([]jira.Ticket, error) {
	var tickets []jira.Ticket
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return nil, fmt.Errorf("could not retrieve users bucket from bolt")
	}
	err = b.ForEach(func(k, v []byte) error {
		var ticket jira.Ticket
		err := json.Unmarshal(v, &ticket)
		if err == nil {
			tickets = append(tickets, ticket)
		}
		return err
	})
	return tickets, err
}

// Slice returns a ticket slice given a low and high bound.
func (db *BoltDB) Slice(l, h int) ([]jira.Ticket, error) {
	if l >= h {
		return nil, fmt.Errorf("low bound is greater than high bound")
	}
	if l < 0 || h < 0 {
		return nil, fmt.Errorf("bounds are negative")
	}
	size, err := db.Size()
	if err != nil {
		return nil, err
	}
	if l > size || h > size {
		return nil, fmt.Errorf("bounds greater than bucket size")
	}
	tickets := make([]jira.Ticket, h-l)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		cursor := b.Cursor()
		_, v := cursor.First()
		var i int
		for i < l {
			_, v = cursor.Next()
			i++
		}
		for i < h {
			var ticket jira.Ticket
			err := json.Unmarshal(v, &ticket)
			if err != nil {
				return err
			}
			tickets[i-l] = ticket
			_, v = cursor.Next()
			i++
		}
		return nil
	})
	return tickets, err
}

// Cursor returns a cursor to the users inside the bucket as well as a function to close the open tx.
func (db *BoltDB) Cursor() (*bolt.Cursor, func() error, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, nil, err
	}
	b := tx.Bucket([]byte(bucketName))
	teardown := func() error {
		return tx.Rollback()
	}
	return b.Cursor(), teardown, nil
}

// Size returns the total number of key/value pairs inside the tickets bucket.
func (db *BoltDB) Size() (int, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()
	return tx.Bucket([]byte(bucketName)).Stats().KeyN, nil
}
