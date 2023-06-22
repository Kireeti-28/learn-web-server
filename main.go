package main

import (
	"net/http"
	"time"
)

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

func main() {
	// Create Server Mux
	mux := http.NewServeMux()
	// Middleware function that adds CORS headers
	corsMux := middlewareCors(mux)
	// Specify address
	const addr = ":8080"

	server := http.Server{
		Handler: corsMux,
		Addr: addr,
		WriteTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	}
	// start server
	server.ListenAndServe()
}