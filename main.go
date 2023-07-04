package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kireeti-28/learn-web-server/internal/database"
)

type apiConfig struct {
	fileserverHits int
	DB             *database.DB
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

/*
	func healthzHandler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("Ok"))
	}

	func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg.fileserverHits = cfg.fileserverHits + 1

			next.ServeHTTP(w, r)
		})
	}

	func (cfg *apiConfig) metricHandler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		resp := fmt.Sprintf(`<html>

<body>

	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>

</body>

</html>

		`, cfg.fileserverHits)
		w.Write([]byte(resp))
		fmt.Printf("Hits: %d\n", cfg.fileserverHits)
	}

	type reqBody struct {
		Body string `json:"body"`
	}

	type errorRespBody struct {
		Error string `json:"error"`
	}

	type validRespBody struct {
		Valid bool `json:"valid"`
	}

	type cleanRespBody struct {
		Cleaned_body string `json:"cleaned_body"`
	}
*/

func (cfg *apiConfig) getChripsIdHandler(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []database.Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, database.Chirp{
			ID:    dbChirp.ID,
			Email: dbChirp.Email,
		})
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	for _, dat := range chirps {
		if dat.ID == id {
			respondWithJSON(w, http.StatusOK, dat)
			return
		}
	}

	respondWithError(w, http.StatusNotFound, "not found id")
}

func (cfg *apiConfig) getChripsHandler(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []database.Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, database.Chirp{
			ID:    dbChirp.ID,
			Email: dbChirp.Email,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) postChripHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.DB.CreateChirp(cleaned)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, database.Chirp{
		ID:    chirp.ID,
		Email: chirp.Email,
	})

}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func (cfg *apiConfig) postUserHandler(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	req := &user{}
	err := decoder.Decode(req)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	db, err := cfg.DB.CreateChirp(req.Email)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	respondWithJSON(w, http.StatusCreated, database.Chirp{
		ID:    db.ID,
		Email: db.Email,
	})

}

func main() {
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	cfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
	}
	r := chi.NewRouter()
	r.Use(middlewareCors)

	r.Get("/api/chirps", cfg.getChripsHandler)
	r.Post("/api/chirps", cfg.postChripHandler)
	r.Get("/api/chirps/{id}", cfg.getChripsIdHandler)

	r.Post("/api/users", cfg.postUserHandler)

	// Specify address
	const addr = ":8080"

	server := http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	// start server
	fmt.Println("Starting server on port ", addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
