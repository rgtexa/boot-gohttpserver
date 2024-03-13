package godb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
)

type Chirp struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
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

	err := db.ensureDB()
	if err != nil {
		log.Fatal(err)
	}

	return &DB{path: path, mux: &sync.Mutex{}}
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	chirpsData, err := db.loadDB()
	if err != nil {
		log.Fatal(err, " loadDB in CreateChirp")
	}
	if len(chirpsData.Chirps) == 0 {
		chirpsData.Chirps = make(map[int]Chirp)
	}
	chirp := Chirp{
		ID:   len(chirpsData.Chirps) + 1,
		Body: body,
	}
	chirpsData.Chirps[chirp.ID] = chirp
	err = db.writeDB(chirpsData)
	if err != nil {
		log.Fatal(err, " writeDB in CreateChirp")
	}
	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	chirpsData, err := db.loadDB()
	if err != nil {
		log.Fatal(err, " loadDB in GetChirps")
	}
	chirps := []Chirp{}
	for chirp := range chirpsData.Chirps {
		chirps = append(chirps, chirpsData.Chirps[chirp])
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID < chirps[j].ID })
	return chirps, nil
}

func (db *DB) GetChirpByID(id int) (Chirp, error) {
	chirpsData, err := db.loadDB()
	if err != nil {
		log.Fatal(err, " loadDB in GetChirpByID")
	}
	chirp, ok := chirpsData.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("Chirp not found")
	}
	return chirp, nil
}

func (db *DB) ensureDB() error {
	db.mux.Lock()
	_, err := os.Create(db.path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.mux.Unlock()
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	chirpsFile, err := os.ReadFile(db.path)
	if err != nil {
		log.Fatal(err, " ReadFile in loadDB")
	}
	chirps := DBStructure{}
	if len(chirpsFile) == 0 {
		return chirps, nil
	}
	err = json.Unmarshal(chirpsFile, &chirps)
	var tt *json.SyntaxError
	if errors.As(err, &tt) {
		fmt.Print(err, " Unmarshal in loadDB")
	}
	return chirps, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	f, err := os.OpenFile(db.path, os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err, " OpenFile in writeDB")
		return err
	}
	chirpjson, err := json.Marshal(dbStructure)
	if err != nil {
		log.Fatal(err, " Marshal in writeDB")
		return err
	}
	_, err = f.WriteAt([]byte(chirpjson), 0)
	if err != nil {
		log.Fatal(err, " WriteAt in writeDB")
	}
	defer f.Close()
	defer db.mux.Unlock()
	return nil
}
