package godb

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type Chirp struct {
	ID   int    `json:"id"`
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
	db := DB{path: path, mux: &sync.Mutex{}}
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		err = db.ensureDB()
		if err != nil {
			log.Fatal(err)
		}
	}
	return &DB{path: path, mux: &sync.Mutex{}}
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	chirp := Chirp{
		Body: body,
	}
	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	chirpsData, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	chirps := []Chirp{}
	for chirp := range chirpsData.Chirps {
		chirps = append(chirps, chirpsData.Chirps[chirp])
	}
	return chirps, nil
}

func (db *DB) ensureDB() error {
	_, err := os.Create(db.path)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	chirpsFile, err := os.ReadFile(db.path)
	if err != nil {
		log.Fatal(err)
	}
	chirps := DBStructure{}
	err = json.Unmarshal(chirpsFile, &chirps)
	if err != nil {
		log.Fatal(err)
	}
	return chirps, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	f, err := os.OpenFile(db.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	defer db.mux.Unlock()
	return nil
}
