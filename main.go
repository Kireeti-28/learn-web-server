package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
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

func respondWithError(w http.ResponseWriter, statusCode int, body errorRespBody) {
	dat, _ := json.Marshal(body)
	w.WriteHeader(statusCode)
	w.Write(dat)
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := reqBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respBody := errorRespBody{
			Error: "Something went wrong",
		}
		respondWithError(w, http.StatusInternalServerError, respBody)
		return
	}

	if len(params.Body) > 140 {
		respBody := errorRespBody{
			Error: "Chirp is too long",
		}
		respondWithError(w, http.StatusBadRequest, respBody)
		return
	}

	respBody := validRespBody{
		Valid: true,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		respBody := errorRespBody{
			Error: "Something went wrong",
		}
		respondWithError(w, http.StatusInternalServerError, respBody)
		return
	}

	loweredText := strings.ToLower(params.Body)
	loweredTextSlice := strings.Split(loweredText, " ")

	cleanTextSlice := []string{}

	for _, word := range(loweredTextSlice) {
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			cleanTextSlice = append(cleanTextSlice, "****")
		} else {
			cleanTextSlice = append(cleanTextSlice, word)
		}
	}

	ans := strings.Join(cleanTextSlice, " ")

	cleanResp := cleanRespBody{
		Cleaned_body: ans,
	}

	dat, _ = json.Marshal(cleanResp)

	w.WriteHeader(200)
	w.Write(dat)
}

func main() {

	r := chi.NewRouter()
	r.Use(middlewareCors)

	r.Post("/api/validate_chirp", validateChirpHandler)

	// config := apiConfig{}
	// handler := config.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	// r.Handle("/app/*", handler)
	// r.Handle("/app", handler)

	// apiRouter := chi.NewRouter()

	// apiRouter.Get("/healthz", healthzHandler)
	// apiRouter.Get("/metrics", config.metricHandler)

	// r.Mount("/admin", apiRouter)

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
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
