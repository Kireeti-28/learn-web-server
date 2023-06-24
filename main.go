package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
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
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits)))
	fmt.Printf("Hits: %d\n", cfg.fileserverHits)
}

func main() {
	mux := http.NewServeMux()
	corsMux := middlewareCors(mux)

	config := apiConfig{}

	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("/healthz", healthzHandler)

	mux.HandleFunc("/metrics", config.metricHandler)

	// Specify address
	const addr = ":8080"

	server := http.Server{
		Handler:      corsMux,
		Addr:         addr,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	// start server
	fmt.Println("Starting server on port ", addr)
	err := server.ListenAndServe()
	log.Fatal(err)
}
