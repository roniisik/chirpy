package main

import (
	"fmt"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type:", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerFileserverHitsCount(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type:", "text/html")
	count := cfg.fileserverHits.Load()
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, count)))
}

func (cfg *apiConfig) handlerFileserverHitsReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Access forbidden", nil)
	}
	cfg.dbQueries.ResetUsers(r.Context())
	cfg.fileserverHits.Add(-cfg.fileserverHits.Load())
}
