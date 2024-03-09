package godb

import (
	"sync"
)

type Chirp struct {
	Body string `json:"body"`
}

type ChirpResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

type DB struct {
	path string
	mux  *sync.Mutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) *DB {
	return &DB{path: path, mux: &sync.Mutex{}}
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	return Chirp{
		Body: body,
	}, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	return []Chirp{}, nil
}

func (db *DB) ensureDB() error {
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	return DBStructure{}, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	return nil
}
