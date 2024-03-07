package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
}

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
		//rctx := chi.RouteContext(r.Context())
		//pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
		fs.ServeHTTP(w, r)
	})
	api.Get("/healthz", handlerReadiness)
	api.Get("/reset", apiCfg.handlerReset)
	api.Post("/validate_chirp", validateChirp)
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

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}
	type chirpErr struct {
		Error string `json:"error"`
	}

	type chirpValid struct {
		Valid bool `json:"valid"`
	}
	decoder := json.NewDecoder(r.Body)
	c := chirp{}
	err := decoder.Decode(&c)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "%v", chirpErr{"Something went wrong"})
		return
	}
	if len(c.Body) > 140 {
		w.WriteHeader(400)
		fmt.Fprintf(w, "%v", chirpErr{"Chirp is too long"})
		return
	}
	respBody := chirpValid{Valid: true}
	dat, err := json.Marshal(respBody)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error marshalling JSON: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}
