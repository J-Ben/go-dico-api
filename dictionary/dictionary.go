package dictionary

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

var (
	bucketName = []byte("dictionary")
)

// Entry représente un mot et sa définition dans le dictionnaire.
type Entry struct {
	Word       string    `json:"word"`
	Definition string    `json:"definition"`
	CreatedAt  time.Time `json:"created_at"`
}

// Dictionary représente un dictionnaire.
type Dictionary struct {
	db *bolt.DB
}

// NewDictionary crée une nouvelle instance de Dictionary en utilisant la base de données spécifiée.
func NewDictionary(dbPath string) (*Dictionary, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Dictionary{
		db: db,
	}, nil
}

// Close ferme la base de données du dictionnaire.
func (d *Dictionary) Close() error {
	return d.db.Close()
}

// AddWord ajoute un mot avec sa définition au dictionnaire.
func (d *Dictionary) AddWord(word, definition string, createdAt time.Time) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return errors.New("dictionary bucket not found")
		}

		entry := Entry{
			Word:       word,
			Definition: definition,
			CreatedAt:  createdAt,
		}

		entryJSON, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(word), entryJSON)
	})
}

// GetWord récupère la définition d'un mot du dictionnaire.
func (d *Dictionary) GetWord(word string) (Entry, error) {
	var entry Entry

	err := d.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return errors.New("dictionary bucket not found")
		}

		entryJSON := bucket.Get([]byte(word))
		if entryJSON == nil {
			return errors.New("word not found")
		}

		return json.Unmarshal(entryJSON, &entry)
	})

	return entry, err
}

// DeleteWord supprime un mot du dictionnaire.
func (d *Dictionary) DeleteWord(word string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return errors.New("dictionary bucket not found")
		}

		return bucket.Delete([]byte(word))
	})
}

// GetAllWords récupère tous les mots du dictionnaire.
func (d *Dictionary) GetAllWords() ([]Entry, error) {
	var entries []Entry

	err := d.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return errors.New("dictionary bucket not found")
		}

		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var entry Entry
			if err := json.Unmarshal(value, &entry); err != nil {
				return err
			}
			entries = append(entries, entry)
		}

		return nil
	})

	return entries, err
}
