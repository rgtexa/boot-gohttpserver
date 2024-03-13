package main

import (
	"encoding/json"
	"fmt"
	"internal/godb"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
}

const dbFile = "chirps.json"

var db = godb.NewDB(dbFile)

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	r := chi.NewRouter()
	api := chi.NewRouter()
	admin := chi.NewRouter()

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)
	r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
		fs.ServeHTTP(w, r)
	})
	api.Get("/healthz", handlerReadiness)
	api.Get("/reset", apiCfg.handlerReset)
	api.Post("/chirps", validateChirp)
	api.Get("/chirps", getChirps)
	api.Get("/chirps/{id}", getChirpByID)
	api.Post("/users", createUser)
	api.Get("/users", getUsers)
	api.Get("/users/{id}", getUserByID)
	r.Mount("/api", api)

	admin.Get("/metrics", apiCfg.handlerMetrics)
	r.Mount("/admin", admin)

	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error marshalling JSON: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func cleanChirp(s string) string {
	words := []string{"kerfuffle", "Kerfuffle", "fornax", "Fornax", "sharbert", "Sharbert"}
	for _, word := range words {
		s = strings.ReplaceAll(s, word, "****")
	}
	return s
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	c := godb.Chirp{}
	err := decoder.Decode(&c)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if len(c.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	c.Body = cleanChirp(c.Body)

	chirp, err := db.CreateChirp(c.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func getChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func getChirpByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	chirp, err := db.GetChirpByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, chirp)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	u := godb.User{}
	err := decoder.Decode(&u)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	user, err := db.CreateUser(u.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	respondWithJSON(w, http.StatusCreated, user)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := db.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	respondWithJSON(w, http.StatusOK, users)
}

func getUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	user, err := db.GetUserByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
	}
	respondWithJSON(w, http.StatusOK, user)
}
