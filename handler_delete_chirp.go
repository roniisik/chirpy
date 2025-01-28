package main

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/roniisik/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "missing access token in header", nil)
		return
	} else if strings.HasPrefix(authHeader, "Bearer ") == false {
		respondWithError(w, http.StatusBadRequest, "missing 'Bearer ' in header", nil)
		return
	}
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	user_id, err := auth.ValidateJWT(accessToken, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "token not valid", err)
		return
	}

	splitPath := strings.Split(r.URL.Path, "/")
	chirpIDStr := splitPath[len(splitPath)-1]
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error parsing wildcard", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "record not found", err)
		return
	}
	if chirp.UserID != user_id {
		respondWithError(w, http.StatusForbidden, "user id not matching", err)
		return
	}

	err = cfg.dbQueries.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed delete query", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}
